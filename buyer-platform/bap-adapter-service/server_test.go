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

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/pubsubtest"
)

func TestInitServerSuccess(t *testing.T) {
	const (
		projectID = "test-project"
		topicID   = "test-topic"
		subID     = "test-subscription"
	)
	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{
			TopicID: topicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  subID,
					Filter: "attributes.action=on_search",
				},
			},
		},
	}
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	httpClient := http.DefaultClient
	conf := config.BuyerAdapterConfig{
		ProjectID:      projectID,
		BuyerAppURL:    "buyer.com/api",
		SubscriptionID: []string{subID},
	}

	_, err = initServer(ctx, httpClient, pubsubClient, conf)

	if err != nil {
		t.Errorf("initServer() failed: %v", err)
	}
}

func TestInitServerFailed(t *testing.T) {
	const (
		projectID = "test-project"
		topicID   = "test-topic"
		subID     = "test-subscription"
	)
	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{
			TopicID: topicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  subID,
					Filter: "attributes.action=on_search",
				},
			},
		},
	}
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		httpClient *http.Client
		psClient   *pubsub.Client
		config     config.BuyerAdapterConfig
	}{
		{
			httpClient: nil,
		},
		{
			httpClient: http.DefaultClient,
			psClient:   nil,
		},
		{
			httpClient: http.DefaultClient,
			psClient:   pubsubClient,
			config: config.BuyerAdapterConfig{
				SubscriptionID: []string{"non-exist-topic"},
			},
		},
	}
	for _, test := range tests {
		_, err = initServer(ctx, test.httpClient, test.psClient, test.config)

		if err == nil { // If NO error
			t.Error("initServer() success unexpectedly")
		}
	}
}

func TestServSuccess(t *testing.T) {
	const (
		projectID = "test-project"
		topicID   = "test-topic"
		subID     = "test-subscription"
	)
	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{
			TopicID: topicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  subID,
					Filter: "attributes.action=on_search",
				},
			},
		},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetups)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	mockBuyerApp := initMockBuyerApp(t)
	httpClient := http.DefaultClient
	conf := config.BuyerAdapterConfig{
		ProjectID:      projectID,
		BuyerAppURL:    mockBuyerApp.URL,
		SubscriptionID: []string{subID},
	}

	srv, err := initServer(ctx, httpClient, pubsubClient, conf)
	if err != nil {
		t.Errorf("initServer() failed: %v", err)
	}

	// publish multiple messages with different action attributes to the topic
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	data := []byte("Hello World")
	var mIDs []string
	for _, act := range []string{"on_search", "on_select"} {
		attrs := map[string]string{"action": act}
		mID := psSrv.Publish(fullTopicID, data, attrs)
		mIDs = append(mIDs, mID)
	}

	// 1 second should be more than enough to handle some messages before canceling the operation.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := srv.serve(ctx); err != nil {
		t.Fatalf("serve() failed: %v", err)
	}

	for _, mID := range mIDs {
		if psSrv.Message(mID).Acks == 0 {
			t.Errorf("Message %q: got no ack", mID)
		}
	}
}

func initMockBuyerApp(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	for _, path := range []string{"/on_search", "/on_select"} {
		mux.HandleFunc(path, func(http.ResponseWriter, *http.Request) {})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
