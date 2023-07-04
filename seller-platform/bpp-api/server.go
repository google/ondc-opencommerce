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

// Server serves HTTP requests as a BPP in the ONDCnetwork.
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
	transactionClient *transactionclient.Client
	topic             *pubsub.Topic
	mux               http.Handler
	conf              config.BPPAPIConfig
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.BPPAPIConfig](configPath)
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

	srv, err := initServer(ctx, conf, registryClient, pubsubClient, transactionClient, clock.New())
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

func initServer(ctx context.Context, conf config.BPPAPIConfig, registryClient middleware.RegistryClient, pubsubClient *pubsub.Client, transactionClient *transactionclient.Client, clk clock.Clock) (*server, error) {
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
		transactionClient: transactionClient,
		topic:             topic,
		conf:              conf,
	}

	mux := http.NewServeMux()
	// Search requests are from the ONDC gateway.
	// Need to authenticate the authorization header of the gateway.
	wrappedSearchHandler := middleware.Adapt(
		http.HandlerFunc(srv.searchHandler),
		middleware.BGAuthentication(registryClient, clk, errorcode.RoleSellerApp, conf.SubscriberID),
	)
	mux.Handle("/search", wrappedSearchHandler)

	for _, e := range [9]struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/select", srv.selectHandler},
		{"/init", srv.initHandler},
		{"/confirm", srv.confirmHandler},
		{"/status", srv.statusHandler},
		{"/track", srv.trackHandler},
		{"/cancel", srv.cancelHandler},
		{"/update", srv.updateHandler},
		{"/rating", srv.ratingHandler},
		{"/support", srv.supportHandler},
	} {
		mux.HandleFunc(e.path, e.handler)
	}

	srv.mux = middleware.Adapt(
		mux,
		middleware.NPAuthentication(registryClient, clk, errorcode.RoleSellerApp, conf.SubscriberID),
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
func nackResponse(w http.ResponseWriter, errorType, errorCode string) {
	res := model.AckResponse{
		Message: &model.MessageAck{
			Ack: &model.Ack{
				Status: "NACK",
			},
		},
		Error: &model.Error{
			Type: errorType,
			Code: &errorCode,
		},
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resJSON)
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.conf.Port)
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
func genericHandler[R model.BPPRequest](s *server, action string, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Errorf("Read request body: %v", err)
		return
	}

	var payload R
	if err := decodeAndValidate(body, &payload); err != nil {
		log.Errorf("Request body is invalid: %v", err)
		msgContext := payload.GetContext()
		errType := "JSON-SCHEMA-ERROR"
		errCode, ok := errorcode.Lookup(errorcode.RoleSellerApp, errorcode.ErrInvalidRequest)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		errCodeStr := strconv.Itoa(errCode)

		if err := s.storeInvalidTransaction(ctx, action, payload, msgContext, errType, errCodeStr, err.Error()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Errorf("Store transaction failed: %v", err)
			return
		}

		nackResponse(w, errType, errCodeStr)
		return
	}

	msgID, err := s.publishMessage(ctx, body, action)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Errorf("Publish Pub/Sub message failed: %v", err)
		return
	}
	w.Header().Set(psMsgIDHeader, msgID)

	if err := s.storeValidTransaction(ctx, action, payload, payload.GetContext()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Errorf("Store transaction failed: %v", err)
		return
	}
	ackResponse(w)
}

func (s *server) storeValidTransaction(ctx context.Context, action string, payload any, msgContext model.Context) error {
	transactionData := transactionclient.TransactionData{
		ID:              *msgContext.TransactionID,
		Type:            "REQUEST-ACTION",
		API:             action,
		MessageID:       *msgContext.MessageID,
		Payload:         payload,
		ProviderID:      *msgContext.BapID,
		MessageStatus:   "ACK",
		ReqReceivedTime: time.Now(),
	}
	return s.transactionClient.StoreTransaction(ctx, transactionData)
}

func (s *server) storeInvalidTransaction(ctx context.Context, action string, payload any, msgContext model.Context, errorType, errorCode, errMsg string) error {
	transactionData := transactionclient.TransactionData{
		ID:              *msgContext.TransactionID,
		Type:            "REQUEST-ACTION",
		API:             action,
		MessageID:       *msgContext.MessageID,
		Payload:         payload,
		ProviderID:      *msgContext.BapID,
		MessageStatus:   "NACK",
		ErrorCode:       errorCode,
		ErrorType:       errorType,
		ErrorMessage:    errMsg,
		ReqReceivedTime: time.Now(),
	}
	return s.transactionClient.StoreTransaction(ctx, transactionData)
}

func (s *server) searchHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.SearchRequest](s, "search", w, r)
}

func (s *server) selectHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.SelectRequest](s, "select", w, r)
}

func (s *server) initHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.InitRequest](s, "init", w, r)
}

func (s *server) confirmHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.ConfirmRequest](s, "confirm", w, r)
}

func (s *server) statusHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.StatusRequest](s, "status", w, r)
}

func (s *server) trackHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.TrackRequest](s, "track", w, r)
}

func (s *server) cancelHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.CancelRequest](s, "cancel", w, r)
}

func (s *server) updateHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.UpdateRequest](s, "update", w, r)
}

func (s *server) ratingHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.RatingRequest](s, "rating", w, r)
}

func (s *server) supportHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.SupportRequest](s, "support", w, r)
}
