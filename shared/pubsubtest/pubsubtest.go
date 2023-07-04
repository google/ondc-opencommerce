// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package pubsubtest provide utility function for testing services that interact with pub/sub.
package pubsubtest

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PubsubSetup represents a state of the pub/sub topics and subscriptions.
type PubsubSetup struct {
	TopicID   string
	SubSetups []SubSetup
}

// SubSetup represents a data for creating a subscription.
type SubSetup struct {
	SubID  string
	Filter string
}

// InitServer create a fake pubsub server, its connection option and setup the pub/sub env.
func InitServer(t *testing.T, projectID string, setups []PubsubSetup) (*pstest.Server, option.ClientOption) {
	t.Helper()
	ctx := context.Background()

	srv := pstest.NewServer()
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("Could not cleanup test Pub/Sub server: %v", err)
		}
	})

	// Connect to the server without using TLS.
	// No need to directly close this connection, it will be closed by closing Pub/Sub client.
	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Use the connection when creating a pubsub client.
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("Could not cleanup Pub/Sub client: %v", err)
		}
	})

	// Create pubsub topics and subscriptions for testing.
	for _, setup := range setups {
		topic, err := client.CreateTopic(ctx, setup.TopicID)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		for _, subSetup := range setup.SubSetups {
			subConfig := pubsub.SubscriptionConfig{
				Topic:  topic,
				Filter: subSetup.Filter,
			}
			_, err := client.CreateSubscription(ctx, subSetup.SubID, subConfig)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
		}
	}

	return srv, option.WithGRPCConn(conn)
}
