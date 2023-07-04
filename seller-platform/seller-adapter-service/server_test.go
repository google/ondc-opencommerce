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

func TestInitializeServerFail(t *testing.T) {
	const (
		projectID       = "test-project"
		bppTopicID      = "bpp-topic"
		callbackTopicID = "callback-topic"
		bppSubID        = "bpp-subscription"
	)
	ctx := context.Background()
	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID: bppTopicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  bppSubID,
					Filter: "attributes.action=search",
				},
			},
		},
		{
			TopicID: callbackTopicID,
		},
	}

	_, opt := pubsubtest.InitServer(t, projectID, psSetup)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		httpClient *http.Client
		conf       config.SellerAdapterConfig
	}{
		{
			httpClient: nil,
		},
		{
			httpClient: http.DefaultClient,
		},
		{
			httpClient: http.DefaultClient,
			conf: config.SellerAdapterConfig{
				ProjectID:       projectID,
				SellerSystemURL: "fakeseller.com/api",
				CallbackTopicID: "non-existent-topic",
				SubscriptionID:  []string{bppSubID},
			},
		},
	}

	for _, test := range tests {
		_, err := initializeServer(ctx, test.httpClient, pubsubClient, test.conf)

		if err == nil { // If NO error
			t.Errorf("initializeServer() success unexpectedly.")
		}
	}
}

func TestInitializeServerSuccess(t *testing.T) {
	const (
		projectID       = "test-project"
		bppTopicID      = "bpp-topic"
		callbackTopicID = "callback-topic"
		bppSubID        = "bpp-subscription"
	)
	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{
			TopicID: bppTopicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  bppSubID,
					Filter: "attributes.action=search",
				},
			},
		},
		{
			TopicID: callbackTopicID,
		},
	}

	httpClient := http.DefaultClient
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		conf config.SellerAdapterConfig
	}{
		{
			conf: config.SellerAdapterConfig{
				ProjectID:       projectID,
				SellerSystemURL: "seller.com/api",
				CallbackTopicID: callbackTopicID,
				SubscriptionID:  []string{bppSubID},
			},
		},
	}

	for _, test := range tests {
		_, err := initializeServer(ctx, httpClient, pubsubClient, test.conf)
		if err != nil {
			t.Errorf("initializeServer() failed: %v", err)
		}
	}
}

func TestHandleSubscriptionCallSellerFail(t *testing.T) {
	const (
		projectID       = "test-project"
		bppTopicID      = "bpp-topic"
		callbackTopicID = "callback-topic"
		bppSubID        = "bpp-subscription"
		action          = "search"
	)
	ctx := context.Background()
	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID: bppTopicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  bppSubID,
					Filter: fmt.Sprintf("attributes.action=%s", action),
				},
			},
		},
		{
			TopicID: callbackTopicID,
		},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetup)

	mockSellerServer := initializeTestSellerSystemServer(t)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	httpClient := mockSellerServer.Client()
	conf := config.SellerAdapterConfig{
		ProjectID:       projectID,
		SellerSystemURL: "",
		CallbackTopicID: callbackTopicID,
		SubscriptionID:  []string{bppSubID},
	}
	srv, err := initializeServer(ctx, httpClient, pubsubClient, conf)
	if err != nil {
		t.Fatalf("initializeServer failed: %v", err)
	}
	// publish a new message for testing.
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", projectID, bppTopicID)
	data := []byte("Hello World")
	attrs := map[string]string{"action": action, "callback_url": "artrary.com/api"}
	_ = psSrv.Publish(fullTopicID, data, attrs)

	// 1 second should be more than enough to handle some messages before canceling the operation.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_ = srv.handleSubscription(ctx, srv.subs[0])
}

func TestHandleSubscriptionReceivingFail(t *testing.T) {
	const (
		projectID       = "test-project"
		bppTopicID      = "bpp-topic"
		callbackTopicID = "callback-topic"
		bppSubID        = "bpp-subscription"
		action          = "search"
	)
	ctx := context.Background()
	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID: bppTopicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  bppSubID,
					Filter: fmt.Sprintf("attributes.action=%s", action),
				},
			},
		},
		{
			TopicID: callbackTopicID,
		},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetup)

	mockSellerServer := initializeTestSellerSystemServer(t)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	httpClient := mockSellerServer.Client()
	conf := config.SellerAdapterConfig{
		ProjectID:       projectID,
		SellerSystemURL: mockSellerServer.URL,
		CallbackTopicID: callbackTopicID,
		SubscriptionID:  []string{bppSubID},
	}
	srv, err := initializeServer(ctx, httpClient, pubsubClient, conf)
	if err != nil {
		t.Fatalf("initializeServer failed: %v", err)
	}
	// publish a new message for testing.
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", projectID, bppTopicID)
	data := []byte("Hello World")
	attrs := map[string]string{"callback_url": "artrary.com/api"}
	_ = psSrv.Publish(fullTopicID, data, attrs)

	// 1 second should be more than enough to handle some messages before canceling the operation.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_ = srv.handleSubscription(ctx, srv.subs[0])
}

func TestHandleSubscriptionSuccess(t *testing.T) {
	const (
		projectID       = "test-project"
		bppTopicID      = "bpp-topic"
		callbackTopicID = "callback-topic"
		bppSubID        = "bpp-subscription"
		action          = "search"
	)
	ctx := context.Background()
	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID: bppTopicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  bppSubID,
					Filter: fmt.Sprintf("attributes.action=%s", action),
				},
			},
		},
		{
			TopicID: callbackTopicID,
		},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetup)

	mockSellerServer := initializeTestSellerSystemServer(t)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	httpClient := mockSellerServer.Client()
	conf := config.SellerAdapterConfig{
		ProjectID:       projectID,
		SellerSystemURL: mockSellerServer.URL,
		CallbackTopicID: callbackTopicID,
		SubscriptionID:  []string{bppSubID},
	}
	srv, err := initializeServer(ctx, httpClient, pubsubClient, conf)
	if err != nil {
		t.Fatalf("initializeServer failed: %v", err)
	}

	// publish a new message for testing.
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", projectID, bppTopicID)
	data := []byte("Hello World")
	attrs := map[string]string{"action": action, "callback_url": "artrary.com/api"}
	mID := psSrv.Publish(fullTopicID, data, attrs)

	// 1 second should be more than enough to handle some messages before canceling the operation.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := srv.handleSubscription(ctx, srv.subs[0]); err != nil {
		t.Errorf("handleSubscription() failed: %v", err)
	}

	if psSrv.Message(mID).Acks == 0 {
		t.Errorf("Message %q: got no ack", mID)
	}
}

func TestServeSuccess(t *testing.T) {
	var (
		projectID       = "test-project"
		bppTopicID      = "bpp-topic"
		callbackTopicID = "callback-topic"
		bppSubSetups    = []pubsubtest.SubSetup{
			{
				SubID:  "bpp-search-subscription",
				Filter: "attributes.action=search",
			},
			{
				SubID:  "bpp-init-subscription",
				Filter: "attributes.action=init",
			},
		}
	)
	ctx := context.Background()

	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID:   bppTopicID,
			SubSetups: bppSubSetups,
		},
		{
			TopicID: callbackTopicID,
		},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetup)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	mockSellerServer := initializeTestSellerSystemServer(t)

	var bppSubIDs []string
	for _, subSetup := range bppSubSetups {
		bppSubIDs = append(bppSubIDs, subSetup.SubID)
	}

	httpClient := mockSellerServer.Client()
	conf := config.SellerAdapterConfig{
		ProjectID:       projectID,
		SellerSystemURL: mockSellerServer.URL,
		CallbackTopicID: callbackTopicID,
		SubscriptionID:  bppSubIDs,
	}

	srv, err := initializeServer(ctx, httpClient, pubsubClient, conf)
	if err != nil {
		t.Fatalf("initializeServer failed: %v", err)
	}

	// publish multiple messages with different action attributes to the topic
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", projectID, bppTopicID)
	data := []byte("Hello World")
	var mIDs []string
	for _, act := range []string{"search", "init"} {
		attrs := map[string]string{"action": act, "callback_url": "artrary.com/api"}
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

// initializeTestSellerSystemServer creates a mockup seller server for testing specifically.
func initializeTestSellerSystemServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(http.ResponseWriter, *http.Request) {})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
