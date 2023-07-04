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

package registryclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/registry"
)

const publicSigningKey = "uJPNW/7CA0iSwMrNbkBAMkVAOoclIIN9lrSeZ2zUkP4="

func TestNewSuccess(t *testing.T) {
	url := "https://preprod.registry.ondc.org/ondc"
	if _, err := New(url, ""); err != nil {
		t.Errorf("New(%q) failed: %v", url, err)
	}
}

func TestNewFailed(t *testing.T) {
	url := `%+o`
	if _, err := New(url, ""); err == nil { // If NO error
		t.Errorf("New(%q) succeeded unexpectedly", url)
	}
}

func TestPublicSigningKey(t *testing.T) {
	publicSigningKeyByte, err := base64.StdEncoding.DecodeString(publicSigningKey)
	if err != nil {
		t.Fatalf("base64.StdEncoding.DecodeString(%q) failed: %v", publicSigningKey, err)
	}
	mockRegistrySrv := initMockRegistryServer(t)
	c, err := New(mockRegistrySrv.URL, "")
	if err != nil {
		t.Fatalf("New(%q) failed: %v", mockRegistrySrv.URL, err)
	}

	country := "IMD"
	city := "Mumbai"
	ondcContext := model.Context{
		Domain:  &model.Domain{Value: "domain"},
		Country: &country,
		City:    &city,
	}
	key, err := c.PublicSigningKey("id", "id", ondcContext)
	if err != nil {
		t.Fatalf("PublicSigningKey() failed: %v", err)
	}
	if bytes.Compare(key, publicSigningKeyByte) != 0 {
		t.Errorf("PublicSigningKey() = %q, want %q", key, publicSigningKey)
	}
}

func TestRotateKeys(t *testing.T) {
	mockRegistrySrv := initMockRegistryServer(t)
	c, err := New(mockRegistrySrv.URL, "")
	if err != nil {
		t.Fatalf("New(%q) failed: %v", mockRegistrySrv.URL, err)
	}

	if err := c.RotateKeys("", "", "", "", time.Second); err != nil {
		t.Errorf("RotateKeys() failed: %v", err)
	}
}

func initMockRegistryServer(t *testing.T) *httptest.Server {
	t.Helper()

	subResponse := registry.SubscribeResponse{
		Message: &registry.SubscribeResponseMessage{
			Ack: &registry.Ack{Status: "ACK"},
		},
	}
	subResponseJSON, err := json.Marshal(subResponse)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/lookup", func(w http.ResponseWriter, _ *http.Request) {
		resp := fmt.Sprintf(`[{"signing_public_key": "%s"}]`, publicSigningKey)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	})
	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(subResponseJSON)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
