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

package middleware

import (
	"encoding/base64"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/registryclienttest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/errorcode"
)

// Valid test case data for authentication middlewares
const (
	testPayload          = `{"context":{"domain":"nic2004:60212","country":"IND","city":"Kochi","action":"search","core_version":"0.9.1","bap_id":"bap.stayhalo.in","bap_uri":"https://8f9f-49-207-209-131.ngrok.io/protocol/","transaction_id":"e6d9f908-1d26-4ff3-a6d1-3af3d3721054","message_id":"a2fe6d52-9fe4-4d1a-9d0b-dccb8b48522d","timestamp":"2022-01-04T09:17:55.971Z","ttl":"P1M"},"message":{"intent":{"fulfillment":{"start":{"location":{"gps":"10.108768, 76.347517"}},"end":{"location":{"gps":"10.102997, 76.353480"}}}}}}`
	testSigningPublicKey = "awGPjRK6i/Vg/lWr+0xObclVxlwZXvTjWYtlu6NeOHk="
	testAuthHeader       = `Signature keyId="example-bap.com|bap1234|ed25519",algorithm="ed25519",created="1641287875",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`
	testCurrentTimestamp = 1641290000
)

var testEmptyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func TestAdapt(t *testing.T) {
	// this middleware append string to the `Test-Header` header.
	testAdapter := func(val string) Adapter {
		return func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				currentHeader := w.Header().Get("Test-Header")
				w.Header().Set("Test-Header", currentHeader+val)
				handler.ServeHTTP(w, r)
			})
		}
	}
	testHandler := Adapt(
		testEmptyHandler,
		testAdapter("1"),
		testAdapter("2"),
		testAdapter("3"),
	)

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	response := httptest.NewRecorder()

	testHandler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusOK; got != want {
		t.Errorf("Status: got %d, want %d", got, want)
	}
	if got, want := response.Header().Get("Test-Header"), "321"; got != want {
		t.Errorf("Header: got %q, want %q", got, want)
	}
}

func TestNPAuthenticationSuccess(t *testing.T) {
	stubRegistryClient, mockClock := createMocksForAuthMiddleware(t, testSigningPublicKey, testCurrentTimestamp)
	testHandler := Adapt(
		testEmptyHandler,
		NPAuthentication(stubRegistryClient, mockClock, errorcode.RoleSellerApp, "bpp.com"),
	)

	request := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(testPayload))
	request.Header.Set("Authorization", testAuthHeader)
	response := httptest.NewRecorder()

	testHandler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusOK; got != want {
		t.Errorf("Status: got %d, want %d", got, want)
	}
	if got, want := response.Header().Get("WWW-Authenticate"), ""; got != want {
		t.Errorf("WWW-Authenticate Header: got %q, want %q", got, want)
	}
}

func TestBGAuthenticationSuccess(t *testing.T) {
	stubRegistryClient, mockClock := createMocksForAuthMiddleware(t, testSigningPublicKey, testCurrentTimestamp)
	testHandler := Adapt(
		testEmptyHandler,
		BGAuthentication(stubRegistryClient, mockClock, errorcode.RoleSellerApp, "bpp.com"),
	)

	request := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(testPayload))
	request.Header.Set("X-Gateway-Authorization", testAuthHeader)
	response := httptest.NewRecorder()

	testHandler.ServeHTTP(response, request)

	if got, want := response.Code, http.StatusOK; got != want {
		t.Errorf("Status: got %d, want %d", got, want)
	}
	if got, want := response.Header().Get("Proxy-Authenticate"), ""; got != want {
		t.Errorf("Proxy-Authenticate Header: got %q, want %q", got, want)
	}
}

func TestNPAuthenticationFail(t *testing.T) {
	tests := []struct {
		payload          string
		pubSigningKey    string
		authHeader       string
		currentTimestamp int
	}{
		{
			payload:          testPayload,
			pubSigningKey:    testSigningPublicKey,
			authHeader:       `Signature`,
			currentTimestamp: testCurrentTimestamp,
		},
		{
			payload:          testPayload,
			pubSigningKey:    testSigningPublicKey,
			authHeader:       `Signature keyId="example-bap.com|bap1234|x25519",algorithm="ed25519",created="1641287875",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`,
			currentTimestamp: testCurrentTimestamp,
		},
		{
			payload:          testPayload,
			pubSigningKey:    testSigningPublicKey,
			authHeader:       testAuthHeader,
			currentTimestamp: 1541287875,
		},
		{
			payload:          testPayload,
			pubSigningKey:    testSigningPublicKey,
			authHeader:       testAuthHeader,
			currentTimestamp: 1741287875,
		},
		{
			payload:          "invalid json",
			pubSigningKey:    testSigningPublicKey,
			authHeader:       testAuthHeader,
			currentTimestamp: testCurrentTimestamp,
		},
		{
			payload:          testPayload,
			pubSigningKey:    "awGPjRK6i/Vg/lWr+0xObclVxlwZXvTjWYtlu7NeOHk=",
			authHeader:       testAuthHeader,
			currentTimestamp: testCurrentTimestamp,
		},
		{
			payload:          testPayload,
			pubSigningKey:    "",
			authHeader:       testAuthHeader,
			currentTimestamp: testCurrentTimestamp,
		},
	}

	for _, test := range tests {
		stubRegistryClient, mockClock := createMocksForAuthMiddleware(t, test.pubSigningKey, test.currentTimestamp)
		testHandler := Adapt(
			testEmptyHandler,
			NPAuthentication(stubRegistryClient, mockClock, errorcode.RoleSellerApp, "bpp.com"),
		)

		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.payload))
		request.Header.Set("Authorization", test.authHeader)
		response := httptest.NewRecorder()

		testHandler.ServeHTTP(response, request)

		if got, want := response.Code, http.StatusUnauthorized; got != want {
			t.Errorf("Status: got %d, want %d", got, want)
		}
		if got, want := response.Header().Get("WWW-Authenticate"), `Signature realm="bpp.com",headers="(created) (expires) digest"`; got != want {
			t.Errorf("WWW-Authenticate Header: got %q, want %q", got, want)
		}
	}
}

func TestOnlyPostMethod(t *testing.T) {
	testHandler := Adapt(testEmptyHandler, OnlyPostMethod())
	tests := []struct {
		method     string
		wantStatus int
	}{
		{
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
		}, {
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		request := httptest.NewRequest(tc.method, "/search", strings.NewReader(""))
		response := httptest.NewRecorder()

		testHandler.ServeHTTP(response, request)

		if got, want := response.Code, tc.wantStatus; got != want {
			t.Errorf("Status: got %d, want %d", got, want)
		}
	}
}

func TestLogging(t *testing.T) {
	flag.Set("v", "1")
	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	})
	testHandler := Adapt(echoHandler, Logging())
	tests := []struct {
		body string
	}{
		{body: "Test Body"},
		{body: ""},
	}

	for _, tc := range tests {
		request := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(tc.body))
		response := httptest.NewRecorder()

		testHandler.ServeHTTP(response, request)

		// test that the request body is not lost due to logging.
		if got, want := response.Body.String(), tc.body; got != want {
			t.Errorf("Response body: got %q, want %q", got, want)
		}
	}
}

func createMocksForAuthMiddleware(t *testing.T, signingPublicKey string, timestamp int) (*registryclienttest.Stub, *clock.Mock) {
	t.Helper()

	signingKey, err := base64.StdEncoding.DecodeString(signingPublicKey)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if len(signingKey) == 0 {
		signingKey = nil
	}
	stubRegistryClient := registryclienttest.NewStub()
	stubRegistryClient.SetKey(signingKey)

	mockClock := clock.NewMock()
	mockClock.Set(time.Unix(int64(timestamp), 0))

	return stubRegistryClient, mockClock
}
