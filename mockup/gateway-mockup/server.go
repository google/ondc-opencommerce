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

// Server handles requests as a gateway in ONDC network.
package main

import (
	"bytes"
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

	"github.com/benbjohnson/clock"
	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/registryclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/errorcode"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/middleware"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/signing-authentication/authentication"
)

var validate = model.Validator()

type server struct {
	conf      config.MockGatewayConfig
	mux       http.Handler
	clk       clock.Clock
	keyClient keyClient
}

type keyClient interface {
	ServiceSigningPrivateKeyset(context.Context) ([]byte, error)
}

type request interface {
	model.SearchRequest | model.OnSearchRequest
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.MockGatewayConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	registryClient, err := registryclient.New(conf.RegistryURL, conf.ONDCEnvironment)
	if err != nil {
		log.Exit(err)
	}

	keyClient, err := keyclient.New(ctx, conf.ProjectID, conf.SecretID)
	if err != nil {
		log.Exit(err)
	}

	srv := initServer(conf, keyClient, registryClient, clock.New())
	log.Info("Server initialization successs")

	err = srv.serve()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server is closed")
	} else if err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initServer(conf config.MockGatewayConfig, keyClient keyClient, registryClient middleware.RegistryClient, clk clock.Clock) *server {
	srv := &server{conf: conf, clk: clk, keyClient: keyClient}

	mux := http.NewServeMux()
	mux.HandleFunc("/search", srv.searchHandler)
	mux.HandleFunc("/on_search", srv.onSearchHandler)
	srv.mux = middleware.Adapt(
		mux,
		middleware.NPAuthentication(registryClient, clk, errorcode.RoleSellerApp, conf.SubscriberID),
		middleware.OnlyPostMethod(),
		middleware.Logging(),
	)

	return srv
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.conf.Port)
	log.Info("Server is serving")
	return http.ListenAndServe(addr, s.mux)
}

func genericHandler[R request](s *server, action string, urls []string, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Read request body: %v", err)
		return
	}

	var payload R
	if err := decodeAndValidate(body, &payload); err != nil {
		nackResponse(w)
		log.Errorf("Request body is invalid: %v", err)
		return
	}

	authHeader := r.Header.Get("Authorization")
	requests, err := s.createONDCRequests(ctx, action, authHeader, urls, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Create requests failed: %v", err)
		return
	}

	for _, request := range requests {
		res, err := http.DefaultClient.Do(request)
		if err != nil {
			log.Errorf("Sending request to %s failed: %v", request.URL, err)
			continue
		}
		resRaw, _ := io.ReadAll(res.Body)
		log.Infof("Sending request to %s: status code %d, body %s", request.URL, res.StatusCode, resRaw)
	}

	ackResponse(w)
}

// createONDCRequests create a HTTP request for ONDC network with a Authorization header.
func (s *server) createONDCRequests(ctx context.Context, action, authHeader string, urls []string, body []byte) ([]*http.Request, error) {
	keyset, err := s.keyClient.ServiceSigningPrivateKeyset(ctx)
	if err != nil {
		return nil, err
	}

	currentTime := s.clk.Now()
	expiredTime := currentTime.Add(3 * time.Minute)
	gatewayAuthHeader, err := authentication.CreateAuthSignature(body, keyset, currentTime.Unix(), expiredTime.Unix(), s.conf.SubscriberID, s.conf.KeyID)
	if err != nil {
		return nil, err
	}

	requests := make([]*http.Request, 0, len(urls))
	for _, url := range urls {
		request, err := http.NewRequest(http.MethodPost, url+action, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		request.Header.Set("X-Gateway-Authorization", gatewayAuthHeader)
		request.Header.Set("Authorization", authHeader)
		request.Header.Set("Content-Type", "application/json")

		requests = append(requests, request)
	}

	return requests, nil
}

func (s *server) searchHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.SearchRequest](s, "/search", s.conf.BPPURLs, w, r)
}

func (s *server) onSearchHandler(w http.ResponseWriter, r *http.Request) {
	genericHandler[model.OnSearchRequest](s, "/on_search", s.conf.BAPURLs, w, r)
}

// decodeAndValidate decodes JSON body and validate the payload.
func decodeAndValidate(body []byte, payload any) error {
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	return validate.Struct(payload)
}

func nackResponse(w http.ResponseWriter) {
	errCode, ok := errorcode.Lookup(errorcode.RoleGateway, errorcode.ErrInvalidRequest)
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
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(resJSON)
}

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
