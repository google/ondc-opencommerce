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

// Server provides Open Commerce API.
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

	"cloud.google.com/go/pubsub"
	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/errorcode"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/middleware"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
)

const psMsgIDHeader = "Pubsub-Message-ID"

var validate = model.Validator()

type server struct {
	pubsubClient *pubsub.Client
	topic        *pubsub.Topic
	mux          http.Handler
	conf         config.BuyerAppConfig
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.BuyerAppConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	pubsubClient, err := pubsub.NewClient(ctx, conf.ProjectID)
	if err != nil {
		log.Exit(err)
	}

	srv, err := initServer(ctx, conf, pubsubClient)
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

func initServer(ctx context.Context, conf config.BuyerAppConfig, pubsubClient *pubsub.Client) (*server, error) {
	topic := pubsubClient.Topic(conf.TopicID)
	exist, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("init server: checking if the topic %q exists: %v", conf.TopicID, err)
	}
	if !exist {
		return nil, fmt.Errorf("init server: topic %q does not exist", conf.TopicID)
	}

	srv := &server{
		pubsubClient: pubsubClient,
		topic:        topic,
		conf:         conf,
	}

	mux := http.NewServeMux()
	apis := [10]struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/search", srv.searchHandler},
		{"/select", srv.selectHandler},
		{"/init", srv.initHandler},
		{"/confirm", srv.confirmHandler},
		{"/status", srv.statusHandler},
		{"/track", srv.trackHandler},
		{"/cancel", srv.cancelHandler},
		{"/update", srv.updateHandler},
		{"/rating", srv.ratingHandler},
		{"/support", srv.supportHandler},
	}
	for _, api := range apis {
		mux.HandleFunc(api.path, api.handler)
	}

	srv.mux = middleware.Adapt(
		mux,
		middleware.OnlyPostMethod(),
		middleware.Logging(),
	)

	return srv, nil
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.conf.Port)
	log.Info("Server is serving")
	return http.ListenAndServe(addr, s.mux)
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
		nackResponse(w)
		log.Errorf("Request body is invalid: %v", err)
		return
	}

	msgID, err := s.publishMessage(ctx, body, action)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Errorf("Publish Pub/Sub message: %v", err)
		return
	}
	w.Header().Set(psMsgIDHeader, msgID)

	ackResponse(w)
}

// decodeAndValidate decodes JSON body and validate the payload.
func decodeAndValidate(body []byte, payload any) error {
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	return validate.Struct(payload)
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

// nackResponse returns an appropriate status code and response body for invalid request body.
func nackResponse(w http.ResponseWriter) {
	errCode, ok := errorcode.Lookup(errorcode.RoleSellerApp, errorcode.ErrInvalidRequest)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	errCodeStr := strconv.Itoa(errCode)

	res := model.AckResponse{
		Message: &model.MessageAck{
			Ack: &model.Ack{
				Status: "NACK",
			},
		},
		Error: &model.Error{
			Type: "JSON-SCHEMA-ERROR",
			Code: &errCodeStr,
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
