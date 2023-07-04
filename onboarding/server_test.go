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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclienttest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/cryptotest"
)

func TestOnSubscribeHandler(t *testing.T) {
	encryptExample := cryptotest.NewExample(t)
	keyClient := keyclienttest.NewStubWithKeys(t, nil, encryptExample.X25519PrivateKey, encryptExample.X25519PublicKey)
	conf := config.OnboardingConfig{
		RegistryEncryptPubKey: cryptotest.ExamplePublicKeyDERB64,
	}
	srv, err := initServer(keyClient, conf)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		requestMethod      string
		requestBody        string
		responseStatusCode int
		responseBody       string
	}{
		{
			requestMethod:      http.MethodGet,
			requestBody:        "",
			responseStatusCode: http.StatusMethodNotAllowed,
			responseBody:       "",
		},
		{
			requestMethod:      http.MethodPost,
			requestBody:        fmt.Sprintf(`{"subscriber_id":"opencommerce.com", "challenge": "%s"}`, encryptExample.EncryptedText),
			responseStatusCode: http.StatusOK,
			responseBody:       fmt.Sprintf(`{"answer":"%s"}`, encryptExample.PlainText),
		},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.requestMethod, "/on_subscribe", strings.NewReader(test.requestBody))
		response := httptest.NewRecorder()

		srv.onSubscribeHandler(response, request)

		if got, want := response.Code, test.responseStatusCode; got != want {
			t.Errorf("Status Code got %v, want %v", got, want)
		}
		if got, want := response.Body.String(), test.responseBody; got != want {
			t.Errorf("Body got %v, want %v", got, want)
		}
	}
}

func TestSiteVerificationHandler(t *testing.T) {
	conf := config.OnboardingConfig{
		RegistryEncryptPubKey: cryptotest.ExamplePublicKeyDERB64,
	}
	srv, err := initServer(keyclienttest.NewStub(t), conf)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/ondc-site-verification.html", nil)
	response := httptest.NewRecorder()

	srv.siteVerificationHandler(response, request)

	if got, want := response.Code, http.StatusOK; got != want {
		t.Errorf("Status Code got %v, want %v", got, want)
	}
	if got, want := response.Header().Get("Content-Type"), "text/html"; !strings.Contains(got, want) {
		t.Errorf("Content-Type got %v, want %v", got, want)
	}
}
