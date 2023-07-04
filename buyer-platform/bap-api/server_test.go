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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"cloud.google.com/go/pubsub"
	"github.com/benbjohnson/clock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/registryclienttest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/transactionclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/pubsubtest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/transactiontest"

	_ "embed"
)

var (
	//go:embed testdata/ack_response.json
	ackResponsePayload []byte
	//go:embed testdata/nack_response.json
	nackResponsePayload []byte

	//go:embed testdata/invalid_request_template.json
	invalidRequestPayload string

	//go:embed testdata/on_search_request.json
	onSearchRequestPayload []byte
	//go:embed testdata/on_select_request.json
	onSelectRequestPayload []byte
	//go:embed testdata/on_init_request.json
	onInitRequestPayload []byte
	//go:embed testdata/on_confirm_request.json
	onConfirmRequestPayload []byte
	//go:embed testdata/on_status_request.json
	onStatusRequestPayload []byte
	//go:embed testdata/on_track_request.json
	onTrackRequestPayload []byte
	//go:embed testdata/on_cancel_request.json
	onCancelRequestPayload []byte
	//go:embed testdata/on_update_request.json
	onUpdateRequestPayload []byte
	//go:embed testdata/on_rating_request.json
	onRatingRequestPayload []byte
	//go:embed testdata/on_support_request.json
	onSupportRequestPayload []byte
)

var invalidRequestTemplaate = template.Must(template.New("invalid_request").Parse(invalidRequestPayload))

func TestInitServerSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bap-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	conf := config.BAPAPIConfig{
		ProjectID: projectID,
		TopicID:   topicID,
	}
	stubRegClient := registryclienttest.NewStub()
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if _, err := initServer(ctx, conf, pubsubClient, stubRegClient, transactionClient, clock.New()); err != nil {
		t.Errorf("initServer() failed: %v", err)
	}
}

func TestInitServerFailed(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bap-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	validRegClient := registryclienttest.NewStub()
	validPsClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	validTransactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	realClock := clock.New()

	tests := []struct {
		conf              config.BAPAPIConfig
		pubsubClient      *pubsub.Client
		registryClient    *registryclienttest.Stub
		transactionClient *transactionclient.Client
	}{
		{
			pubsubClient: nil,
		},
		{
			pubsubClient:   validPsClient,
			registryClient: nil,
		},
		{
			pubsubClient:      validPsClient,
			registryClient:    validRegClient,
			transactionClient: nil,
		},
		{
			conf: config.BAPAPIConfig{
				ProjectID: projectID,
				TopicID:   "non-exist-topic",
			},
			pubsubClient:      validPsClient,
			registryClient:    validRegClient,
			transactionClient: validTransactionClient,
		},
	}
	for _, test := range tests {
		_, err := initServer(ctx, test.conf, test.pubsubClient, test.registryClient, test.transactionClient, realClock)
		if err == nil { // If NO error
			t.Errorf("initServer() succeeded unexpectedly")
		}
	}
}

func TestHandlersSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bap-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetups)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	conf := config.BAPAPIConfig{
		ProjectID: projectID,
		TopicID:   topicID,
	}
	stubRegClient := registryclienttest.NewStub()
	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	srv, err := initServer(ctx, conf, pubsubClient, stubRegClient, transactionClient, clock.New())
	if err != nil {
		t.Fatalf("initServer() failed: %v", err)
	}

	tests := [10]struct {
		handlerName string
		handler     http.HandlerFunc
		path        string
		body        []byte
	}{
		{
			handlerName: "onSearchHandler",
			handler:     srv.onSearchHandler,
			path:        "/on_search",
			body:        onSearchRequestPayload,
		},
		{
			handlerName: "onSelectHandler",
			handler:     srv.onSelectHandler,
			path:        "/on_select",
			body:        onSelectRequestPayload,
		},
		{
			handlerName: "onInitHandler",
			handler:     srv.onInitHandler,
			path:        "/on_init",
			body:        onInitRequestPayload,
		},
		{
			handlerName: "onConfirmHandler",
			handler:     srv.onConfirmHandler,
			path:        "/on_confirm",
			body:        onConfirmRequestPayload,
		},
		{
			handlerName: "onStatusHandler",
			handler:     srv.onStatusHandler,
			path:        "/on_status",
			body:        onStatusRequestPayload,
		},
		{
			handlerName: "onTrackHandler",
			handler:     srv.onTrackHandler,
			path:        "/on_track",
			body:        onTrackRequestPayload,
		},
		{
			handlerName: "onCancelHandler",
			handler:     srv.onCancelHandler,
			path:        "/on_cancel",
			body:        onCancelRequestPayload,
		},
		{
			handlerName: "onUpdateHandler",
			handler:     srv.onUpdateHandler,
			path:        "/on_update",
			body:        onUpdateRequestPayload,
		},
		{
			handlerName: "onRatingHandler",
			handler:     srv.onRatingHandler,
			path:        "/on_rating",
			body:        onRatingRequestPayload,
		},
		{
			handlerName: "onSupportHandler",
			handler:     srv.onSupportHandler,
			path:        "/on_support",
			body:        onSupportRequestPayload,
		},
	}
	var wantAck model.AckResponse
	if err := json.Unmarshal(ackResponsePayload, &wantAck); err != nil {
		t.Fatalf("Unmarshal want response body got error: %v", err)
	}

	for _, test := range tests {
		test := test // Make a local copy of test data for safety.
		t.Run(test.handlerName, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodPost, test.path, bytes.NewReader(test.body))
			response := httptest.NewRecorder()

			test.handler(response, request)

			if got, want := response.Code, http.StatusOK; got != want {
				t.Logf("%s: response body: %s", test.handlerName, response.Body.String())
				t.Fatalf("%s got status %d, want %d", test.handlerName, got, want)
			}

			var gotAck model.AckResponse
			if err := json.Unmarshal(response.Body.Bytes(), &gotAck); err != nil {
				t.Fatalf("%s Unmarshal response body got error: %v", test.handlerName, err)
			}
			if diff := cmp.Diff(wantAck, gotAck); diff != "" {
				t.Errorf("%s response body diff (-want, +got):\n%s", test.handlerName, diff)
			}

			msgID := response.Header().Get(psMsgIDHeader)
			psMsg := psSrv.Message(msgID)
			if psMsg == nil {
				t.Fatalf("%s publish no message", test.handlerName)
			}
			if bytes.Compare(psMsg.Data, test.body) != 0 {
				t.Errorf("%s Pub/Sub message data is not equal to request body", test.handlerName)
			}
		})
	}
}

func TestHandlersInvalidPayload(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bap-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetups)
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	conf := config.BAPAPIConfig{
		ProjectID: projectID,
		TopicID:   topicID,
	}
	stubRegClient := registryclienttest.NewStub()
	srv, err := initServer(ctx, conf, pubsubClient, stubRegClient, transactionClient, clock.New())
	if err != nil {
		t.Fatalf("initServer() failed: %v", err)
	}

	tests := [10]struct {
		handlerName string
		handler     http.HandlerFunc
		path        string
	}{
		{
			handlerName: "onSearchHandler",
			handler:     srv.onSearchHandler,
			path:        "/on_search",
		},
		{
			handlerName: "onSelectHandler",
			handler:     srv.onSelectHandler,
			path:        "/on_select",
		},
		{
			handlerName: "onInitHandler",
			handler:     srv.onInitHandler,
			path:        "/on_init",
		},
		{
			handlerName: "onConfirmHandler",
			handler:     srv.onConfirmHandler,
			path:        "/on_confirm",
		},
		{
			handlerName: "onStatusHandler",
			handler:     srv.onStatusHandler,
			path:        "/on_status",
		},
		{
			handlerName: "onTrackHandler",
			handler:     srv.onTrackHandler,
			path:        "/on_track",
		},
		{
			handlerName: "onCancelHandler",
			handler:     srv.onCancelHandler,
			path:        "/on_cancel",
		},
		{
			handlerName: "onUpdateHandler",
			handler:     srv.onUpdateHandler,
			path:        "/on_update",
		},
		{
			handlerName: "onRatingHandler",
			handler:     srv.onRatingHandler,
			path:        "/on_rating",
		},
		{
			handlerName: "onSupportHandler",
			handler:     srv.onSupportHandler,
			path:        "/on_support",
		},
	}
	var wantAck model.AckResponse
	if err := json.Unmarshal(nackResponsePayload, &wantAck); err != nil {
		t.Fatalf("Unmarshal want response body got error: %v", err)
	}

	for _, test := range tests {
		test := test // Make a local copy of test data for safety.
		t.Run(test.handlerName, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			invalidRequestTemplaate.Execute(&buf, uuid.New())
			request := httptest.NewRequest(http.MethodPost, test.path, &buf)
			response := httptest.NewRecorder()

			test.handler(response, request)

			if got, want := response.Code, http.StatusBadRequest; got != want {
				t.Logf("%s: response body: %s", test.handlerName, response.Body.String())
				t.Fatalf("%s got status %d, want %d", test.handlerName, got, want)
			}

			var gotAck model.AckResponse
			if err := json.Unmarshal(response.Body.Bytes(), &gotAck); err != nil {
				t.Fatalf("%s Unmarshal response body got error: %v", test.handlerName, err)
			}
			if diff := cmp.Diff(wantAck, gotAck); diff != "" {
				t.Errorf("%s response body diff (-want, +got):\n%s", test.handlerName, diff)
			}

			msgID := response.Header().Get(psMsgIDHeader)
			psMsg := psSrv.Message(msgID)
			if psMsg != nil {
				t.Errorf("%s unexpectedly publish a message", test.handlerName)
			}
		})
	}
}
