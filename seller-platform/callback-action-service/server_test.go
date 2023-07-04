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
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/benbjohnson/clock"
	"github.com/google/uuid"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclienttest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/transactionclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/pubsubtest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/transactiontest"

	_ "embed"
)

var (
	//go:embed testdata/on_search_request.json
	onSearchRequest string

	//go:embed testdata/on_select_request.json
	onSelectRequest string

	//go:embed testdata/on_init_request.json
	onInitRequest string

	//go:embed testdata/on_confirm_request.json
	onConfirmRequest string

	//go:embed testdata/on_track_request.json
	onTrackRequest string

	//go:embed testdata/on_cancel_request.json
	onCancelRequest string

	//go:embed testdata/on_update_request.json
	onUpdateRequest string

	//go:embed testdata/on_status_request.json
	onStatusRequest string

	//go:embed testdata/on_rating_request.json
	onRatingRequest string

	//go:embed testdata/on_support_request.json
	onSupportRequest string
)

var (
	onSearchReqTemplate  = template.Must(template.New("on_search").Parse(onSearchRequest))
	onSelectReqTemplate  = template.Must(template.New("on_select").Parse(onSelectRequest))
	onInitReqTemplate    = template.Must(template.New("on_init").Parse(onInitRequest))
	onConfirmReqTemplate = template.Must(template.New("on_confirm").Parse(onConfirmRequest))
	onTrackReqTemplate   = template.Must(template.New("on_track").Parse(onTrackRequest))
	onCancelReqTemplate  = template.Must(template.New("on_cancel").Parse(onCancelRequest))
	onUpdateReqTemplate  = template.Must(template.New("on_update").Parse(onUpdateRequest))
	onStatusReqTemplate  = template.Must(template.New("on_status").Parse(onStatusRequest))
	onRatingReqTemplate  = template.Must(template.New("on_rating").Parse(onRatingRequest))
	onSupportReqTemplate = template.Must(template.New("on_support").Parse(onSupportRequest))
)

func TestInitializeServerSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	subID := fmt.Sprintf("callback-subscription-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

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

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	httpClient := http.DefaultClient
	keyClient := keyclienttest.NewStub(t)
	realClock := clock.New()

	tests := []struct {
		conf config.CallbackActionConfig
	}{
		{
			conf: config.CallbackActionConfig{
				ProjectID:      projectID,
				TopicID:        topicID,
				SubscriptionID: []string{subID},
			},
		},
	}

	for _, test := range tests {
		_, err := initServer(ctx, httpClient, pubsubClient, keyClient, transactionClient, test.conf, realClock)
		if err != nil {
			t.Errorf("initServer() failed: %v", err)
		}
	}
}

func TestInitializeServerFail(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	subID := fmt.Sprintf("callback-subscription-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

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

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	realClock := clock.New()

	tests := []struct {
		httpClient *http.Client
		keyClient  keyClient
		conf       config.CallbackActionConfig
	}{
		{
			httpClient: nil,
			keyClient:  nil,
		},
		{
			httpClient: http.DefaultClient,
			keyClient:  nil,
		},
		{
			httpClient: http.DefaultClient,
			keyClient:  keyclienttest.NewStub(t),
		},
		{
			httpClient: http.DefaultClient,
			keyClient:  keyclienttest.NewStub(t),
			conf: config.CallbackActionConfig{
				ProjectID:      projectID,
				TopicID:        "non-exist-topic",
				SubscriptionID: []string{subID},
			},
		},
	}

	for _, test := range tests {
		_, err := initServer(ctx, test.httpClient, pubsubClient, test.keyClient, transactionClient, test.conf, realClock)

		if err == nil { // If NO error
			t.Errorf("initServer() success unexpectedly.")
		}
	}
}

func TestServeSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	subIDs := []string{
		"callback-subscription-on-search",
		"callback-subscription-on-select",
		"callback-subscription-on-init",
		"callback-subscription-on-confirm",
		"callback-subscription-on-track",
		"callback-subscription-on-cancel",
		"callback-subscription-on-update",
		"callback-subscription-on-status",
		"callback-subscription-on-rating",
		"callback-subscription-on-support",
	}
	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID: topicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  "callback-subscription-on-search",
					Filter: "attributes.action=on_search",
				},
				{
					SubID:  "callback-subscription-on-select",
					Filter: "attributes.action=on_select",
				},
				{
					SubID:  "callback-subscription-on-init",
					Filter: "attributes.action=on_init",
				},
				{
					SubID:  "callback-subscription-on-confirm",
					Filter: "attributes.action=on_confirm",
				},
				{
					SubID:  "callback-subscription-on-track",
					Filter: "attributes.action=on_track",
				},
				{
					SubID:  "callback-subscription-on-cancel",
					Filter: "attributes.action=on_cancel",
				},
				{
					SubID:  "callback-subscription-on-update",
					Filter: "attributes.action=on_update",
				},
				{
					SubID:  "callback-subscription-on-status",
					Filter: "attributes.action=on_status",
				},
				{
					SubID:  "callback-subscription-on-rating",
					Filter: "attributes.action=on_rating",
				},
				{
					SubID:  "callback-subscription-on-support",
					Filter: "attributes.action=on_support",
				},
			},
		},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetup)

	mockBAPSrv := initMockBAPServer(t)
	mockGateway := initMockGateway(t)

	ctx := context.Background()
	httpClient := mockBAPSrv.Client()
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	keyClient := keyclienttest.NewStub(t)
	conf := config.CallbackActionConfig{
		ProjectID:      projectID,
		TopicID:        topicID,
		SubscriptionID: subIDs,
		GatewayURL:     mockGateway.URL,
	}
	srv, err := initServer(ctx, httpClient, pubsubClient, keyClient, transactionClient, conf, clock.New())
	if err != nil {
		t.Fatalf("initServer failed: %v", err)
	}

	// publish new messages for testing.
	var mIDs []string
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)

	tests := []struct {
		action      string
		reqTemplate *template.Template
	}{
		{
			action:      "on_search",
			reqTemplate: onSearchReqTemplate,
		},
		{
			action:      "on_select",
			reqTemplate: onSelectReqTemplate,
		},
		{
			action:      "on_init",
			reqTemplate: onInitReqTemplate,
		},
		{
			action:      "on_confirm",
			reqTemplate: onConfirmReqTemplate,
		},
		{
			action:      "on_track",
			reqTemplate: onTrackReqTemplate,
		},
		{
			action:      "on_cancel",
			reqTemplate: onCancelReqTemplate,
		},
		{
			action:      "on_update",
			reqTemplate: onUpdateReqTemplate,
		},
		{
			action:      "on_status",
			reqTemplate: onStatusReqTemplate,
		},
		{
			action:      "on_rating",
			reqTemplate: onRatingReqTemplate,
		},
		{
			action:      "on_support",
			reqTemplate: onSupportReqTemplate,
		},
	}
	for _, test := range tests {
		var data bytes.Buffer
		if err := test.reqTemplate.Execute(&data, mockBAPSrv.URL); err != nil {
			t.Fatal(err)
		}

		attrs := map[string]string{"action": test.action}
		mID := psSrv.Publish(fullTopicID, data.Bytes(), attrs)
		mIDs = append(mIDs, mID)
	}

	// 5 seconds should be more than enough to handle some messages before canceling the operation.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.serve(ctx); err != nil {
		t.Errorf("serve() failed: %v", err)
	}

	for _, mID := range mIDs {
		if psSrv.Message(mID).Acks == 0 {
			t.Errorf("Message %q: got no ack", mID)
		}
	}
}

func ackResponse(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": {"ack": {"status": "ACK"}}}`))
}

func initMockBAPServer(t *testing.T) *httptest.Server {
	t.Helper()

	// For this scenario, on_search should be sent to mock gateway.
	paths := [9]string{
		"/on_select",
		"/on_init",
		"/on_confirm",
		"/on_track",
		"/on_cancel",
		"/on_update",
		"/on_status",
		"/on_rating",
		"/on_support",
	}
	mux := http.NewServeMux()
	for _, path := range paths {
		mux.HandleFunc(path, ackResponse)
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func initMockGateway(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/on_search", ackResponse)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
