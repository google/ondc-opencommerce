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

// Package middleware provides common middleware for services.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/benbjohnson/clock"
	log "github.com/golang/glog"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/errorcode"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	auth "partner-innovation.googlesource.com/googleondcaccelerator.git/shared/signing-authentication/authentication"
)

// Adapter wraps an handler and return a new handler
type Adapter func(handler http.Handler) http.Handler

// RegistryClient provides a public ED25519 key.
type RegistryClient interface {
	PublicSigningKey(subscriberID, uniqueKeyID string, domain model.Context) ([]byte, error)
}

// Adapt wraps the given handler with a list of adapters.
//
// The first adapter of the parameters is the innermost middleware.
func Adapt(handler http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		handler = adapter(handler)
	}
	return handler
}

// OnlyPostMethod returns an error if the method is not POST.
func OnlyPostMethod() Adapter {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				log.Errorf("Invalid HTTP method: got %q, want %q", r.Method, http.MethodPost)
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			// pass the request to handler in case of the POST method
			handler.ServeHTTP(w, r)
		})
	}
}

// Logging is a middleware for logging incoming request detail.
func Logging() Adapter {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Infof("Got a request from %s", r.Host)
			if log.V(1) {
				body, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Infof("Request body:\n%s", body)
			}

			handler.ServeHTTP(w, r)
		})
	}
}

// BGAuthentication is a middleware for authenticating a signature from the Gateway.
func BGAuthentication(registryClient RegistryClient, clock clock.Clock, role errorcode.Role, subscriberID string) Adapter {
	authenticator := &authenticator{
		registryClient:  registryClient,
		clock:           clock,
		role:            role,
		subscriberID:    subscriberID,
		verifyingHeader: "X-Gateway-Authorization",
		nackHeader:      "Proxy-Authenticate",
	}
	return func(handler http.Handler) http.Handler {
		return authenticator.authentication(handler)
	}
}

// NPAuthentication is a middleware for authenticating a signature from ONDC network participants.
func NPAuthentication(registryClient RegistryClient, clock clock.Clock, role errorcode.Role, subscriberID string) Adapter {
	authenticator := &authenticator{
		registryClient:  registryClient,
		clock:           clock,
		role:            role,
		subscriberID:    subscriberID,
		verifyingHeader: "Authorization",
		nackHeader:      "WWW-Authenticate",
	}
	return func(handler http.Handler) http.Handler {
		return authenticator.authentication(handler)
	}
}

type authenticator struct {
	registryClient  RegistryClient
	clock           clock.Clock
	role            errorcode.Role
	subscriberID    string
	verifyingHeader string
	nackHeader      string
}

// authentication is a generic middleware for authenticating a signature from both BG and BAP/BPP.
func (a *authenticator) authentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(a.verifyingHeader)
		info, err := auth.ExtractInfoFromHeader(header)
		if err != nil {
			log.Errorf("Invalid %q header format: %s", a.verifyingHeader, err)
			log.Errorf("Invalid %q header value: %q", a.verifyingHeader, header)
			a.unauthenticated(w)
			return
		}

		if info.Algorithm != info.KeyIDAlgorithm {
			log.Errorf("Invalid %q header: algorithms do not match", a.verifyingHeader)
			log.Errorf("Invalid %q header: algorithm=%q, Key ID algorithm=%q", a.verifyingHeader, info.Algorithm, info.KeyIDAlgorithm)
			a.unauthenticated(w)
			return
		}

		currentTimestamp := a.clock.Now().Unix()
		if info.Created > currentTimestamp || info.Expired < currentTimestamp {
			log.Errorf("Invalid %q header: invalid timestamps: created=%d, expired=%d", a.verifyingHeader, info.Created, info.Expired)
			a.unauthenticated(w)
			return
		}

		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Extract only the ONDC context of the request for key lookup.
		var ondcCtx struct {
			Context model.Context `json:"context"`
		}
		decoder := json.NewDecoder(bytes.NewReader(body))
		if err := decoder.Decode(&ondcCtx); err != nil {
			log.Errorf("Decode context failed: %s", err)
			a.unauthenticated(w)
			return
		}

		ed25519PublicKey, err := a.registryClient.PublicSigningKey(info.SubscriberID, info.UniqueKeyID, ondcCtx.Context)
		if err != nil {
			log.Errorf("Get public signing key from registry failed: %s", err)
			a.unauthenticated(w)
			return
		}

		if err := auth.VerifySignature(info.Signature, body, ed25519PublicKey, info.Created, info.Expired); err != nil {
			log.Errorf("Verify signature failed: %s", err)
			a.unauthenticated(w)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// unauthenticated writes a proper response when the request authentication fails.
func (a *authenticator) unauthenticated(w http.ResponseWriter) {
	errCode, ok := errorcode.Lookup(a.role, errorcode.ErrInvalidSignature)
	if !ok {
		http.Error(w, "", http.StatusInternalServerError)
	}
	errCodeStr := strconv.Itoa(errCode)

	response := model.AckResponse{
		Message: &model.MessageAck{
			Ack: &model.Ack{
				Status: "NACK",
			},
		},
		Error: &model.Error{
			Type: "CONTEXT-ERROR",
			Code: &errCodeStr,
		},
	}
	responseJSON, err := json.Marshal(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	headerValue := fmt.Sprintf(`Signature realm="%s",headers="(created) (expires) digest"`, a.subscriberID)
	w.Header().Set(a.nackHeader, headerValue)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(responseJSON)
}
