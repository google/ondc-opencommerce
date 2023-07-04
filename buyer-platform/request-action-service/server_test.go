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

	"github.com/benbjohnson/clock"
	"github.com/google/uuid"
	"google.golang.org/api/option"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclienttest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/pubsubtest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/transactiontest"

	_ "embed"
)

var (
	//go:embed testdata/search_request.json
	searchRequestPayload string
	//go:embed testdata/select_request.json
	selectRequestPayload string
	//go:embed testdata/init_request.json
	initRequestPayload string
	//go:embed testdata/confirm_request.json
	confirmRequestPayload string
	//go:embed testdata/status_request.json
	statusRequestPayload string
	//go:embed testdata/track_request.json
	trackRequestPayload string
	//go:embed testdata/cancel_request.json
	cancelRequestPayload string
	//go:embed testdata/update_request.json
	updateRequestPayload string
	//go:embed testdata/rating_request.json
	ratingRequestPayload string
	//go:embed testdata/support_request.json
	supportRequestPayload string
)

var (
	searchReqTemplate  = template.Must(template.New("search").Parse(searchRequestPayload))
	selectReqTemplate  = template.Must(template.New("select").Parse(selectRequestPayload))
	initReqTemplate    = template.Must(template.New("init").Parse(initRequestPayload))
	confirmReqTemplate = template.Must(template.New("confirm").Parse(confirmRequestPayload))
	statusReqTemplate  = template.Must(template.New("status").Parse(statusRequestPayload))
	trackReqTemplate   = template.Must(template.New("track").Parse(trackRequestPayload))
	cancelReqTemplate  = template.Must(template.New("cancel").Parse(cancelRequestPayload))
	updateReqTemplate  = template.Must(template.New("update").Parse(updateRequestPayload))
	ratingReqTemplate  = template.Must(template.New("rating").Parse(ratingRequestPayload))
	supportReqTemplate = template.Must(template.New("support").Parse(supportRequestPayload))
)

func TestInitServerSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("test-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	conf := config.RequestActionConfig{
		ProjectID:      projectID,
		SubscriptionID: []string{"test-subscription"},
		InstanceID:     instanceID,
		DatabaseID:     databaseID,
	}

	ctx := context.Background()
	psSetups := []pubsubtest.PubsubSetup{
		{
			TopicID: topicID,
			SubSetups: []pubsubtest.SubSetup{
				{
					SubID:  conf.SubscriptionID[0],
					Filter: "attributes.action=search",
				},
			},
		},
	}
	_, pubsubOpt := pubsubtest.InitServer(t, conf.ProjectID, psSetups)
	transactionOpts := transactiontest.NewDatabase(ctx, t, conf.ProjectID, conf.InstanceID, conf.DatabaseID)
	realClock := clock.New()
	keyClient := keyclienttest.NewStub(t)

	_, err := initServer(ctx, conf, realClock, keyClient, []option.ClientOption{pubsubOpt}, transactionOpts)
	if err != nil {
		t.Errorf("initServer() failed: %v", err)
	}
}

func TestServeSuccess(t *testing.T) {
	hash := uuid.New().String()[:8]
	projectID := fmt.Sprintf("test-project-%s", hash)
	topicID := fmt.Sprintf("test-topic-%s", hash)
	instanceID := fmt.Sprintf("test-instance-%s", hash)
	databaseID := fmt.Sprintf("test-database-%s", hash)

	ctx := context.Background()
	mockBPPSrv := initMockBPPServer(t)

	allActions := []string{"search", "select", "init", "confirm", "status", "track", "cancel", "update", "rating", "support"}
	subscriptionIDs := make([]string, 0, len(allActions))
	subSetups := make([]pubsubtest.SubSetup, 0, len(allActions))
	for _, action := range allActions {
		subID := fmt.Sprintf("test-subscription-%s", action)
		subscriptionIDs = append(subscriptionIDs, subID)
		subSetups = append(subSetups, pubsubtest.SubSetup{
			SubID:  subID,
			Filter: fmt.Sprintf("attributes.action=%s", action),
		})
	}
	conf := config.RequestActionConfig{
		ProjectID:      projectID,
		SubscriptionID: subscriptionIDs,
		InstanceID:     instanceID,
		DatabaseID:     databaseID,
		GatewayURL:     mockBPPSrv.URL,
	}
	psSetup := []pubsubtest.PubsubSetup{
		{
			TopicID:   topicID,
			SubSetups: subSetups,
		},
	}
	psSrv, psOpt := pubsubtest.InitServer(t, conf.ProjectID, psSetup)
	transactionOpts := transactiontest.NewDatabase(ctx, t, conf.ProjectID, conf.InstanceID, conf.DatabaseID)
	realClock := clock.New()
	keyClient := keyclienttest.NewStub(t)
	srv, err := initServer(ctx, conf, realClock, keyClient, []option.ClientOption{psOpt}, transactionOpts)
	if err != nil {
		t.Fatalf("initServer() failed: %v", err)
	}

	// publish new messages for testing.
	var mIDs []string
	fullTopicID := fmt.Sprintf("projects/%s/topics/%s", conf.ProjectID, topicID)
	tests := []struct {
		action      string
		reqTemplate *template.Template
	}{
		{
			action:      "search",
			reqTemplate: searchReqTemplate,
		},
		{
			action:      "select",
			reqTemplate: selectReqTemplate,
		},
		{
			action:      "init",
			reqTemplate: initReqTemplate,
		},
		{
			action:      "confirm",
			reqTemplate: confirmReqTemplate,
		},
		{
			action:      "status",
			reqTemplate: statusReqTemplate,
		},
		{
			action:      "track",
			reqTemplate: trackReqTemplate,
		},
		{
			action:      "cancel",
			reqTemplate: cancelReqTemplate,
		},
		{
			action:      "update",
			reqTemplate: updateReqTemplate,
		},
		{
			action:      "rating",
			reqTemplate: ratingReqTemplate,
		},
		{
			action:      "support",
			reqTemplate: supportReqTemplate,
		},
	}
	for _, test := range tests {
		var data bytes.Buffer
		if err := test.reqTemplate.Execute(&data, mockBPPSrv.URL); err != nil {
			t.Fatal(err)
		}

		attrs := map[string]string{"action": test.action}
		mID := psSrv.Publish(fullTopicID, data.Bytes(), attrs)
		mIDs = append(mIDs, mID)
	}

	// 1 second should be more than enough to handle some messages before canceling the operation.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
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

func initMockBPPServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": {"ack": {"status": "ACK"}}}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
