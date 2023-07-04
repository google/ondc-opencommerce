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

// Server serves HTTP requests for ONDC registry mock-up
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/crypto"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/registry"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	validate              = model.Validator()
)

type keyLookup struct {
	SubscriberID string
	UkID         string
}

type server struct {
	conf       config.MockRegistryConfig
	mux        *http.ServeMux
	keyLookups map[keyLookup]registry.LookupResponseInner
}

func main() {
	flag.Set("alsologtostderr", "true")
	configPath, ok := os.LookupEnv("CONFIG")
	if !ok {
		log.Exit("CONFIG env is not set")
	}

	conf, err := config.Read[config.MockRegistryConfig](configPath)
	if err != nil {
		log.Exit(err)
	}

	srv := initServer(conf)
	log.Info("Server initialization successs")

	err = srv.serve()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server is closed")
	} else if err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initServer(conf config.MockRegistryConfig) *server {
	srv := &server{conf: conf}

	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", srv.subscribeHandler)
	mux.HandleFunc("/lookup", srv.lookupHandler)
	srv.mux = mux

	keyLookups := make(map[keyLookup]registry.LookupResponseInner, len(conf.Keys))
	for _, key := range conf.Keys {
		keyLookups[keyLookup{key.SubscriberID, key.UkID}] = key
	}
	srv.keyLookups = keyLookups

	return srv
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.conf.Port)
	log.Info("Server is serving")
	return http.ListenAndServe(addr, s.mux)
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	var request registry.SubscribeRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		nackResponse(w)
		log.Errorf("Decoding request: %v", err)
		return
	}
	if err := validate.Struct(request); err != nil {
		nackResponse(w)
		log.Errorf("Request body is invalid: %v", err)
		return
	}

	ackResponse(w)

	go s.onSubscribeCallback(request)
}

func (s *server) onSubscribeCallback(request registry.SubscribeRequest) error {
	privKey, err := base64.StdEncoding.DecodeString(s.conf.RegistryKeyset.PrivateEncryptionKey)
	if err != nil {
		log.Errorf("Decode private encryption key failed: %s", err)
		return err
	}

	pubKey, err := base64.StdEncoding.DecodeString(request.Message.Entity.KeyPair.EncryptionPublicKey)
	if err != nil {
		log.Errorf("Decode public encryption key failed: %s", err)
		return err
	}

	chanllenge := randomString(16)
	encryptedChallenge, err := crypto.EncryptMessage(chanllenge, privKey, pubKey)
	if err != nil {
		log.Errorf("Encrypt challenge failed: %s", err)
		return err
	}

	callbackURL := request.Message.Entity.CallbackURL + "/on_subscribe"
	callbackReq := registry.OnSubscribeRequest{
		SubscriberID: request.Message.Entity.SubscriberID,
		Challenge:    encryptedChallenge,
	}
	callbackBodyJSON, err := json.Marshal(callbackReq)
	if err != nil {
		log.Errorf("Marshall request failed: %s", err)
		return err
	}

	callbackRes, err := http.Post(callbackURL, "application/json", bytes.NewReader(callbackBodyJSON))
	if err != nil {
		log.Errorf("Call on_subscribe error: %s", err)
		return err
	}
	defer callbackRes.Body.Close()

	var res registry.OnSubscribeResponse
	if err := json.NewDecoder(callbackRes.Body).Decode(&res); err != nil {
		log.Errorf("Decode response failed: %s", err)
		return err
	}

	if res.Answer != chanllenge {
		errMsg := fmt.Sprintf("Incorrect challenge answer: got %q, want %q", res.Answer, chanllenge)
		log.Info(errMsg)
		return errors.New(errMsg)
	}
	log.Info("The challenge answer is correct")
	return nil
}

func (s *server) lookupHandler(w http.ResponseWriter, r *http.Request) {
	var request registry.LookupRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		log.Errorf("Decode request failed: %s", err)
		nackResponse(w)
		return
	}
	if err := validate.Struct(request); err != nil {
		log.Errorf("Request body is invalid: %v", err)
		nackResponse(w)
		return
	}

	var response registry.LookupResponse
	if key, ok := s.keyLookups[keyLookup{*request.SubscriberID, request.UkID}]; ok {
		response = append(response, key)
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Marshall response failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func ackResponse(w http.ResponseWriter) {
	res := registry.SubscribeResponse{
		Message: &registry.SubscribeResponseMessage{
			Ack: &registry.Ack{
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

func nackResponse(w http.ResponseWriter) {
	res := registry.SubscribeResponse{
		Message: &registry.SubscribeResponseMessage{
			Ack: &registry.Ack{
				Status: "NACK",
			},
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

func randomString(lenght int) string {
	b := make([]byte, lenght)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
