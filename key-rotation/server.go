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

// Server serves HTTP requests for ONDC rotation key
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/registryclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/crypto"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/signing-authentication/authentication"
)

var validate = model.Validator()

// config is a config for key rotation service.
type config struct {
	ProjectID       string
	SecretID        string
	RegistryURL     string
	RequestID       string
	SubscriberID    string
	RotationPeriod  time.Duration
	ONDCEnvironment string
}

// secretManagerEvent is a Pub/Sub message describing an event of the Secret Manager.
type secretManagerEvent struct {
	Message struct {
		Attributes       attribute `json:"attributes"`
		Data             string    `json:"data"`
		MessageID        string    `json:"messageId"`
		MessageIDUnder   string    `json:"message_id"`
		PublishTime      string    `json:"publishTime"`
		PublishTimeUnder string    `json:"publish_time"`
	} `json:"message"`

	Subscription string `json:"subscription"`
}

// attribute contains attributes of a message from the Secret Manager.
type attribute struct {
	DataFormat string `json:"dataFormat"`
	EventType  string `json:"eventType"`
	SecretID   string `json:"secretId"`
	Timestamp  string `json:"timestamp"`
}

type keyClient interface {
	AddKey(ctx context.Context, secretID string, payload []byte) error
}

type registryClient interface {
	RotateKeys(encryptionPublicKey, signingPublicKey, requestID, subscriberID string, d time.Duration) error
}

// server servs HTTP requests for key rotation flow
type server struct {
	mux            *http.ServeMux
	keyClient      keyClient
	registryClient registryClient

	conf config
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	projectId, ok := os.LookupEnv("PROJECT_ID")
	if !ok {
		log.Exit("PROJECT_ID env is not set")
	}

	secretId, ok := os.LookupEnv("SECRET_ID")
	if !ok {
		log.Exit("SECRET_ID env is not set")
	}

	registryUrl, ok := os.LookupEnv("REGISTRY_URL")
	if !ok {
		log.Exit("REGISTRY_URL env is not set")
	}

	requestID, ok := os.LookupEnv("REQUEST_ID")
	if !ok {
		log.Exit("REQUEST_ID env is not set")
	}

	subscriberID, ok := os.LookupEnv("SUBSCRIBER_ID")
	if !ok {
		log.Exit("SUBSCRIBER_ID env is not set")
	}

	rotationPeriod, ok := os.LookupEnv("ROTATION_PERIOD")
	if !ok {
		log.Exit("ROTATION_PERIOD env is not set")
	}
	rotationDuration, err := time.ParseDuration(rotationPeriod)
	if err != nil {
		log.Exitf("ROTATION_PERIOD is invalid: %v", err)
	}

	conf := config{
		ProjectID:      projectId,
		SecretID:       secretId,
		RegistryURL:    registryUrl,
		RequestID:      requestID,
		SubscriberID:   subscriberID,
		RotationPeriod: rotationDuration,
	}

	registryClient, err := registryclient.New(conf.RegistryURL, conf.ONDCEnvironment)
	if err != nil {
		log.Exit(err)
	}

	keyClient, err := keyclient.New(ctx, conf.ProjectID, conf.SecretID)
	if err != nil {
		log.Exit(err)
	}
	defer keyClient.Close()

	srv := initServer(keyClient, registryClient, conf)
	log.Info("Server initialization successs")

	err = srv.serve()
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server is closed")
	} else if err != nil {
		log.Exitf("Serving failed: %v", err)
	}
}

func initServer(keyClient keyClient, registryClient registryClient, conf config) *server {
	server := &server{
		mux:            http.NewServeMux(),
		keyClient:      keyClient,
		registryClient: registryClient,
		conf:           conf,
	}
	server.mux.HandleFunc("/", server.rotationHandler)
	return server
}

func (s *server) serve() error {
	portNumber := os.Getenv("PORT")
	if portNumber == "" {
		portNumber = "8080"
	}
	log.Info("Server is serving")
	return http.ListenAndServe(":"+portNumber, s.mux)
}

func (s *server) rotationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var event secretManagerEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Infof("Decode request body failed: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// The Secret Manager publishes events of various types via Pub/Sub to this service.
	// Only events with the type 'SECRET_ROTATE' should be handled,
	// and all other event types should be ignored.
	eventType := event.Message.Attributes.EventType
	if eventType != "SECRET_ROTATE" {
		msg := fmt.Sprintf("Ignore event type: %q", eventType)
		log.Info(msg)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
		return
	}

	encryptionPrivateKey, encryptionPublicKey, encryptionPublicKeyDER, err := crypto.GenerateEncryptionKeyPair()
	if err != nil {
		log.Errorf("Generate encryption key pair failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signingKeyset, err := authentication.GenerateKeysetJSON()
	if err != nil {
		log.Errorf("Generate signing keyset failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signingPublicKey, err := authentication.ExtractRawPublicKey(signingKeyset)
	if err != nil {
		log.Errorf("Extract raw public signing key failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload := map[string]any{
		"encryptionKey": map[string][]byte{
			"privateKeyEncryption":   encryptionPrivateKey,
			"publicKeyEncryption":    encryptionPublicKey,
			"publicKeyEncryptionDER": encryptionPublicKeyDER,
		},
		"signingKey": map[string][]byte{
			"signingKeySet":    signingKeyset,
			"publicKeySigning": signingPublicKey,
		},
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Marshal keyset payload failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.keyClient.AddKey(ctx, event.Message.Attributes.SecretID, payloadJSON); err != nil {
		log.Errorf("Add key to secret manager failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encryptionPublicKeyB64 := base64.StdEncoding.EncodeToString(encryptionPublicKeyDER)
	signingPublicKeyB64 := base64.StdEncoding.EncodeToString(signingPublicKey)
	reqID := s.conf.RequestID
	subID := s.conf.SubscriberID
	period := s.conf.RotationPeriod
	if err := s.registryClient.RotateKeys(encryptionPublicKeyB64, signingPublicKeyB64, reqID, subID, period); err != nil {
		log.Errorf("Rotate keys in ONDC registry failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completeMsg := "Key rotation is completed"
	log.Info(completeMsg)
	w.Write([]byte(completeMsg))
}
