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

	//go:embed testdata/search_request.json
	searchRequestPayload []byte
	//go:embed testdata/select_request.json
	selectRequestPayload []byte
	//go:embed testdata/init_request.json
	initRequestPayload []byte
	//go:embed testdata/confirm_request.json
	confirmRequestPayload []byte
	//go:embed testdata/status_request.json
	statusRequestPayload []byte
	//go:embed testdata/track_request.json
	trackRequestPayload []byte
	//go:embed testdata/cancel_request.json
	cancelRequestPayload []byte
	//go:embed testdata/update_request.json
	updateRequestPayload []byte
	//go:embed testdata/rating_request.json
	ratingRequestPayload []byte
	//go:embed testdata/support_request.json
	supportRequestPayload []byte
)

func TestInitializeServerSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	ctx := context.Background()
	conf := config.BPPAPIConfig{
		ProjectID: projectID,
		TopicID:   topicID,
	}
	stubRegClient := registryclienttest.NewStub()

	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if _, err := initServer(ctx, conf, stubRegClient, pubsubClient, transactionClient, clock.New()); err != nil {
		t.Errorf("initServer() failed: %v", err)
	}
}

func TestHandlersSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetups)
	ctx := context.Background()
	conf := config.BPPAPIConfig{
		ProjectID: projectID,
		TopicID:   topicID,
	}
	stubRegClient := registryclienttest.NewStub()
	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	srv, err := initServer(ctx, conf, stubRegClient, pubsubClient, transactionClient, clock.New())
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
			handlerName: "searchHandler",
			handler:     srv.searchHandler,
			path:        "/search",
			body:        searchRequestPayload,
		},
		{
			handlerName: "selectHandler",
			handler:     srv.selectHandler,
			path:        "/select",
			body:        selectRequestPayload,
		},
		{
			handlerName: "initHandler",
			handler:     srv.initHandler,
			path:        "/init",
			body:        initRequestPayload,
		},
		{
			handlerName: "confirmHandler",
			handler:     srv.confirmHandler,
			path:        "/confirm",
			body:        confirmRequestPayload,
		},
		{
			handlerName: "statusHandler",
			handler:     srv.statusHandler,
			path:        "/status",
			body:        statusRequestPayload,
		},
		{
			handlerName: "trackHandler",
			handler:     srv.trackHandler,
			path:        "/track",
			body:        trackRequestPayload,
		},
		{
			handlerName: "cancelHandler",
			handler:     srv.cancelHandler,
			path:        "/cancel",
			body:        cancelRequestPayload,
		},
		{
			handlerName: "updateHandler",
			handler:     srv.updateHandler,
			path:        "/update",
			body:        updateRequestPayload,
		},
		{
			handlerName: "ratingHandler",
			handler:     srv.ratingHandler,
			path:        "/rating",
			body:        ratingRequestPayload,
		},
		{
			handlerName: "supportHandler",
			handler:     srv.supportHandler,
			path:        "/support",
			body:        supportRequestPayload,
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
				t.Errorf("%s got status %d, want %d", test.handlerName, got, want)
				t.Logf("Response body: %s", response.Body.Bytes())
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
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	psSrv, opt := pubsubtest.InitServer(t, projectID, psSetups)
	ctx := context.Background()
	conf := config.BPPAPIConfig{
		ProjectID: projectID,
		TopicID:   topicID,
	}
	stubRegClient := registryclienttest.NewStub()
	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	srv, err := initServer(ctx, conf, stubRegClient, pubsubClient, transactionClient, clock.New())
	if err != nil {
		t.Fatalf("initServer() failed: %v", err)
	}

	tests := [10]struct {
		handlerName string
		handler     http.HandlerFunc
		path        string
	}{
		{
			handlerName: "searchHandler",
			handler:     srv.searchHandler,
			path:        "/search",
		},
		{
			handlerName: "selectHandler",
			handler:     srv.selectHandler,
			path:        "/select",
		},
		{
			handlerName: "initHandler",
			handler:     srv.initHandler,
			path:        "/init",
		},
		{
			handlerName: "confirmHandler",
			handler:     srv.confirmHandler,
			path:        "/confirm",
		},
		{
			handlerName: "statusHandler",
			handler:     srv.statusHandler,
			path:        "/status",
		},
		{
			handlerName: "trackHandler",
			handler:     srv.trackHandler,
			path:        "/track",
		},
		{
			handlerName: "cancelHandler",
			handler:     srv.cancelHandler,
			path:        "/cancel",
		},
		{
			handlerName: "updateHandler",
			handler:     srv.updateHandler,
			path:        "/update",
		},
		{
			handlerName: "ratingHandler",
			handler:     srv.ratingHandler,
			path:        "/rating",
		},
		{
			handlerName: "supportHandler",
			handler:     srv.supportHandler,
			path:        "/support",
		},
	}
	var wantAck model.AckResponse
	if err := json.Unmarshal(nackResponsePayload, &wantAck); err != nil {
		t.Fatalf("Unmarshal want response body got error: %v", err)
	}
	invalidTemplate, err := template.New("invalid_request").Parse(invalidRequestPayload)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	for _, test := range tests {
		test := test // Make a local copy of test data for safety.
		t.Run(test.handlerName, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			invalidTemplate.Execute(&buf, uuid.New())
			request := httptest.NewRequest(http.MethodPost, test.path, &buf)
			response := httptest.NewRecorder()

			test.handler(response, request)

			if got, want := response.Code, http.StatusBadRequest; got != want {
				t.Errorf("%s got status %d, want %d", test.handlerName, got, want)
				t.Logf("Response body: %s", response.Body.Bytes())
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

func TestHandlersInitPubSubTopicFail(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("bpp-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	psSetups := []pubsubtest.PubsubSetup{
		{TopicID: topicID},
	}
	_, opt := pubsubtest.InitServer(t, projectID, psSetups)
	ctx := context.Background()
	conf := config.BPPAPIConfig{
		ProjectID: projectID,
		TopicID:   "",
	}
	stubRegClient := registryclienttest.NewStub()
	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID, opt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	opts := transactiontest.NewDatabase(ctx, t, projectID, instanceID, databaseID)
	transactionClient, err := transactionclient.New(ctx, projectID, instanceID, databaseID, opts...)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = initServer(ctx, conf, stubRegClient, pubsubClient, transactionClient, clock.New())
	if err == nil { // If NO error
		t.Fatalf("initServer() succeeded unexpectedly")
	}
}
