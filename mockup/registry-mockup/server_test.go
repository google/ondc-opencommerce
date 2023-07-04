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
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/crypto"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/registry"

	_ "embed"
)

func TestInitServer(t *testing.T) {
	conf := config.MockRegistryConfig{}
	initServer(conf)
}

//go:embed testdata/subscribe_request.json
var subscribeRequestSuccess []byte

func TestSubscribeHandler(t *testing.T) {
	srv := initTestServer(t)
	tests := []struct {
		name           string
		reqBody        []byte
		wantStatusCode int
	}{
		{
			name:           "Valid Request",
			reqBody:        subscribeRequestSuccess,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Empty Request",
			reqBody:        []byte(""),
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		request := httptest.NewRequest(http.MethodPost, "/lookup", bytes.NewReader(test.reqBody))
		response := httptest.NewRecorder()

		srv.subscribeHandler(response, request)

		if got, want := response.Code, test.wantStatusCode; got != want {
			t.Errorf("%s: status code got %d, want %d", test.name, got, want)
		}
	}
}

var (
	//go:embed testdata/lookup_request_success.json
	lookupRequestSuccess []byte
	//go:embed testdata/lookup_request_not_found.json
	lookupResquestNotFound []byte
)

func TestLookupHandler(t *testing.T) {
	srv := initTestServer(t)
	tests := []struct {
		name              string
		reqBody           []byte
		wantStatusCode    int
		wantResponseCount int
	}{
		{
			name:              "Valid Request",
			reqBody:           lookupRequestSuccess,
			wantStatusCode:    http.StatusOK,
			wantResponseCount: 1,
		},
		{
			name:              "Non-exist Subscriber",
			reqBody:           lookupResquestNotFound,
			wantStatusCode:    http.StatusOK,
			wantResponseCount: 0,
		},
	}

	for _, test := range tests {
		request := httptest.NewRequest(http.MethodPost, "/lookup", bytes.NewReader(test.reqBody))
		response := httptest.NewRecorder()

		srv.lookupHandler(response, request)

		if got, want := response.Code, test.wantStatusCode; got != want {
			t.Errorf("%s: status code got %d, want %d", test.name, got, want)
		}
		if got, want := countLookupResponse(t, response), test.wantResponseCount; got != want {
			t.Errorf("%s: response count got %d, want %d", test.name, got, want)
		}
	}
}

func TestOnSubscribeCallbackFail(t *testing.T) {
	mockSubscriberSrv := initMockSubscriberServer(t)

	privKey, _, _, err := crypto.GenerateEncryptionKeyPair()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, pubKey, _, err := crypto.GenerateEncryptionKeyPair()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	conf := config.MockRegistryConfig{
		RegistryKeyset: config.Keyset{
			PrivateEncryptionKey: base64.StdEncoding.EncodeToString(privKey),
		},
	}
	srv := initServer(conf)
	req := registry.SubscribeRequest{
		Message: &registry.SubscribeMessage{
			RequestID: uuid.New().String(),
			Entity: &registry.Entity{
				CallbackURL: mockSubscriberSrv.URL,
				KeyPair: &registry.KeyPair{
					EncryptionPublicKey: base64.StdEncoding.EncodeToString(pubKey),
				},
			},
		},
	}

	err = srv.onSubscribeCallback(req)

	if err == nil { // If NO error
		t.Fatalf("onSubscribeCallback() succeeded unexpectedly")
	}
	if got, want := err.Error(), "Incorrect challenge answer"; !strings.Contains(got, want) {
		t.Errorf("onSubscribeCallback got error %q, want %q", got, want)
	}
}

func countLookupResponse(t *testing.T, res *httptest.ResponseRecorder) int {
	t.Helper()

	var response registry.LookupResponse
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&response); err != nil {
		t.Fatal(err)
	}

	return len(response)
}

func initTestServer(t *testing.T) *server {
	t.Helper()

	conf := config.MockRegistryConfig{
		Port: 8080,
		RegistryKeyset: config.Keyset{
			PublicSigningKey:     "",
			PrivateSigningKey:    "",
			PublicEncryptionKey:  "",
			PrivateEncryptionKey: "",
		},
		Keys: registry.LookupResponse{
			{
				SubscriberID:     "https://sit.grab.in/ondc",
				UkID:             "22a8a67a-76d9-459b-867c-085dda2939ec",
				BrID:             "22a8a67a-76d9-459b-867c-085dda2939ec",
				Country:          "IND",
				City:             "std:080",
				Domain:           "nic2004:52110",
				SigningPublicKey: "awGPjRK6i/Vg/lWr+0xObclVxlwZXvTjWYtlu6NeOHk=",
				EncrPublicKey:    "MCowBQYDK2VuAyEAa9Wbpvd9SsrpOZFcynyt/TO3x0Yrqyys4NUGIvyxX2Q=",
				ValidFrom:        "2022-04-05T05:56:52.470618Z3",
				ValidUntil:       "2026-04-05T05:56:52.470618Z7",
				Created:          "2026-04-05T05:56:52.470618Z7",
				Updated:          "2026-04-05T05:56:52.470618Z7",
			},
		},
	}
	return initServer(conf)
}

func initMockSubscriberServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/on_subscribe", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"answer": "invalid_value"}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
