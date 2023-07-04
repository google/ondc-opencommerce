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

// Server serves HTTP requests for ONDC onboarding process
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
	"strconv"
	"text/template"

	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclient"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/crypto"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/registry"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/signing-authentication/authentication"
)

const siteVerificationHTML = `
<html>
    <head>
        <meta name='ondc-site-verification' content='{{.}}' />
    </head>
    <body>
        ONDC Site Verification Page
    </body>
</html>
`

var siteVerificationTemplate = template.Must(template.New("site-verification").Parse(siteVerificationHTML))

type keyClient interface {
	ServiceSigningPrivateKeyset(context.Context) ([]byte, error)
	ServiceEncryptionPrivateKey(context.Context) ([]byte, error)
}

// server servs HTTP requests for onboarding and subscription flow
type server struct {
	mux       *http.ServeMux
	keyClient keyClient
	conf      config.OnboardingConfig

	registryEncryptPubKey []byte
}

func main() {
	flag.Set("alsologtostderr", "true")
	ctx := context.Background()

	portNumber := os.Getenv("PORT")
	if portNumber == "" {
		portNumber = "8080"
	}
	port, err := strconv.Atoi(portNumber)
	if err != nil {
		log.Exitf("Config errer %v", err)
	}

	projectID, ok := os.LookupEnv("PROJECT_ID")
	if !ok {
		log.Exit("PROJECT_ID env is not set")
	}

	secretID, ok := os.LookupEnv("SECRET_ID")
	if !ok {
		log.Exit("SECRET_ID env is not set")
	}

	requestID, ok := os.LookupEnv("REQUEST_ID")
	if !ok {
		log.Exit("REQUEST_ID env is not set")
	}

	registryEncryptPubKey, ok := os.LookupEnv("REGISTRY_ENCRYPT_PUB_KEY")
	if !ok {
		log.Exit("REGISTRY_ENCRYPT_PUB_KEY env is not set")
	}

	conf := config.OnboardingConfig{
		ProjectID:             projectID,
		Port:                  port,
		RequestID:             requestID,
		SecretID:              secretID,
		RegistryEncryptPubKey: registryEncryptPubKey,
	}

	keyClient, err := keyclient.New(ctx, conf.ProjectID, conf.SecretID)
	if err != nil {
		log.Exit(err)
	}
	defer keyClient.Close()

	srv, err := initServer(keyClient, conf)
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

func initServer(keyClient keyClient, conf config.OnboardingConfig) (*server, error) {
	if keyClient == nil {
		return nil, errors.New("init server: key client is nil")
	}

	pubKeyByte, err := crypto.ExtractRawPubKeyFromDER(conf.RegistryEncryptPubKey)
	if err != nil {
		return nil, fmt.Errorf("init server: invalid registry encryption public key: %v", err)
	}

	server := &server{
		keyClient:             keyClient,
		conf:                  conf,
		registryEncryptPubKey: pubKeyByte,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/on_subscribe", server.onSubscribeHandler)
	mux.HandleFunc("/ondc-site-verification.html", server.siteVerificationHandler)
	server.mux = mux

	return server, nil
}

func (s *server) serve() error {
	addr := fmt.Sprintf(":%d", s.conf.Port)
	log.Info("Server is serving")
	return http.ListenAndServe(addr, s.mux)
}

func (s *server) onSubscribeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var request registry.OnSubscribeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	decryptedMessage, err := s.decryptChallenge(ctx, request.Challenge)
	if err != nil {
		log.Errorf("Decrypt challenge failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := registry.OnSubscribeResponse{
		Answer: decryptedMessage,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func (s *server) decryptChallenge(ctx context.Context, message string) (string, error) {
	privateKey, err := s.keyClient.ServiceEncryptionPrivateKey(ctx)
	if err != nil {
		return "", err
	}

	decryptedMsg, err := crypto.DecryptMessage(message, privateKey, s.registryEncryptPubKey)
	if err != nil {
		return "", err
	}
	return decryptedMsg, nil
}

func (s *server) siteVerificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	signingKeyset, err := s.keyClient.ServiceSigningPrivateKeyset(ctx)
	if err != nil {
		log.Errorf("Failed to fetch signing private key: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signedRequestID, err := authentication.Sign([]byte(s.conf.RequestID), signingKeyset)
	if err != nil {
		log.Errorf("Failed to sign request ID: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signedRequestIDB64 := base64.StdEncoding.EncodeToString(signedRequestID)
	if err = siteVerificationTemplate.Execute(w, signedRequestIDB64); err != nil {
		log.Errorf("Failed to execute template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
