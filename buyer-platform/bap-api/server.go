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

// Server serves HTTP requests as a BAP in the ONDCnetwork.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/benbjohnson/clock"
	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/registryclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/transactionclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/errorcode"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/middleware"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
)

const psMsgIDHeader = "Pubsub-Message-ID"

var validate = model.Validator()

type server struct {
	pubsubClient      *pubsub.Client
	topic             *pubsub.Topic
	mux               http.Handler
	port              int
	transactionClient *transactionclient.Client
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.BAPAPIConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	registryClient, err := registryclient.New(conf.RegistryURL, conf.ONDCEnvironment)
	if err != nil {
		log.Exit(err)
	}

	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Exit(err)
	}

	transactionClient, err := transactionclient.New(ctx, conf.ProjectID, conf.InstanceID, conf.DatabaseID)
	if err != nil {
		log.Exit(err)
	}

	srv, err := initServer(ctx, conf, pubsubClient, registryClient, transactionClient, clock.New())
	if err != nil {
		log.Exit(err)
	}
	log.Info("Server initialization successs")

	err = srv.serve()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server is closed")
	} else if err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initServer(ctx context.Context, conf config.BAPAPIConfig, pubsubClient *pubsub.Client, registryClient middleware.RegistryClient, transactionClient *transactionclient.Client, clk clock.Clock) (*server, error) {
	// validate clients
	if pubsubClient == nil {
		return nil, errors.New("init server: Pub/Sub client is nil")
	}
	if registryClient == nil {
		return nil, errors.New("init server: registry client is nil")
	}
	if transactionClient == nil {
		return nil, errors.New("init server: transaction client is nil")
	}

	topic := pubsubClient.Topic(conf.TopicID)
	exist, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("init server: %v", err)
	}
	if !exist {
		return nil, fmt.Errorf("init server: topic %q does not exist", conf.TopicID)
	}

	srv := &server{
		pubsubClient:      pubsubClient,
		topic:             topic,
		port:              conf.Port,
		transactionClient: transactionClient,
	}

	mux := http.NewServeMux()
	for _, e := range [10]struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/on_search", srv.onSearchHandler},
		{"/on_select", srv.onSelectHandler},
		{"/on_init", srv.onInitHandler},
		{"/on_confirm", srv.onConfirmHandler},
		{"/on_status", srv.onStatusHandler},
		{"/on_track", srv.onTrackHandler},
		{"/on_cancel", srv.onCancelHandler},
		{"/on_update", srv.onUpdateHandler},
		{"/on_rating", srv.onRatingHandler},
		{"/on_support", srv.onSupportHandler},
	} {
		mux.HandleFunc(e.path, e.handler)
	}

	srv.mux = middleware.Adapt(
		mux,
		middleware.NPAuthentication(registryClient, clk, errorcode.RoleBuyerApp, conf.SubscriberID),
		middleware.OnlyPostMethod(),
		middleware.Logging(),
	)

	return srv, nil
}

// decodeAndValidate decodes JSON body and validate the payload.
func decodeAndValidate(body []byte, payload any) error {
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	return validate.Struct(payload)
}

// nackResponse returns an appropriate status code and response body for invalid request body.
func nackResponse(w http.ResponseWriter, errType, errCode string) {
	res := model.AckResponse{
		Message: &model.MessageAck{
			Ack: &model.Ack{
				Status: "NACK",
			},
		},
		Error: &model.Error{
			Type: errType,
			Code: &errCode,
		},
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(resJSON)
}

// ackResponse returns an appropriate status code and response body for valid request body.
func ackResponse(w http.ResponseWriter) {
	res := model.AckResponse{
		Message: &model.MessageAck{
			Ack: &model.Ack{
				Status: "ACK",
			},
		},
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resJSON)
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Info("Server is serving")
	return http.ListenAndServe(addr, s.mux)
}

// publishMessage publishes incoming request to the topic and return the publishing result.
func (s *server) publishMessage(ctx context.Context, body []byte, action string) (msgID string, err error) {
	msg := &pubsub.Message{
		Data: body,
		Attributes: map[string]string{
			"action": action,
		},
	}
	result := s.topic.Publish(ctx, msg)
	return result.Get(ctx)
}

// genericHandler can handles all kind of ONDC request.
func genericHandler[R model.BAPRequest](s *server, action string, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Read request body: %v", err)
		return
	}

	var payload R
	if err := decodeAndValidate(body, &payload); err != nil {
		log.Errorf("Request body is invalid: %v", err)
		errCodeInt, ok := errorcode.Lookup(errorcode.RoleSellerApp, errorcode.ErrInvalidRequest)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		errType := "JSON-SCHEMA-ERROR"
		errCode := strconv.Itoa(errCodeInt)
		if err := s.storeTransaction(ctx, action, "NACK", payload, payload.GetContext(), errType, errCode, err.Error()); err != nil {
			log.Errorf("Store transaction for invalid request failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		nackResponse(w, errType, errCode)
		return
	}

	if err := s.storeTransaction(ctx, action, "ACK", payload, payload.GetContext(), "", "", ""); err != nil {
		log.Errorf("Store transaction for valid request failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	msgID, err := s.publishMessage(ctx, body, action)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Publish Pub/Sub message: %v", err)
		return
	}
	w.Header().Set(psMsgIDHeader, msgID)

	ackResponse(w)
	log.Infof("Successfully ack request: TransactionID: %q, MessageID: %q", *payload.GetContext().TransactionID, *payload.GetContext().MessageID)
}

func (s *server) storeTransaction(ctx context.Context, action, status string, payload any, msgContext model.Context, errType, errCode, errMsg string) error {
	transactionData := transactionclient.TransactionData{
		ID:              *msgContext.TransactionID,
		Type:            "CALLBACK-ACTION",
		API:             action,
		MessageID:       *msgContext.MessageID,
		Payload:         payload,
		ProviderID:      msgContext.BppID,
		MessageStatus:   status,
		ErrorCode:       errCode,
		ErrorType:       errType,
		ErrorMessage:    errMsg,
		ReqReceivedTime: time.Now(),
	}
	return s.transactionClient.StoreTransaction(ctx, transactionData)
}

func (s *server) onSearchHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnSearchRequest](s, "on_search", w, r)
}

func (s *server) onSelectHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnSelectRequest](s, "on_select", w, r)
}

func (s *server) onInitHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnInitRequest](s, "on_init", w, r)
}

func (s *server) onConfirmHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnConfirmRequest](s, "on_confirm", w, r)
}

func (s *server) onStatusHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnStatusRequest](s, "on_status", w, r)
}

func (s *server) onTrackHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnTrackRequest](s, "on_track", w, r)
}

func (s *server) onCancelHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnCancelRequest](s, "on_cancel", w, r)
}

func (s *server) onUpdateHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnUpdateRequest](s, "on_update", w, r)
}

func (s *server) onRatingHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnRatingRequest](s, "on_rating", w, r)
}

func (s *server) onSupportHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnSupportRequest](s, "on_support", w, r)
}
