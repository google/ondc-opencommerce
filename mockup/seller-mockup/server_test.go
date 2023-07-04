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
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/config"

	_ "embed"
)

//go:embed testdata/search_request.json
var testSearchRequest []byte

//go:embed testdata/search_request_uncomplete.json
var testSearchRequestUncomplete []byte

func TestInitServer(t *testing.T) {
	conf := config.MockSellerSystemConfig{Port: 8080}

	_, err := initServer(conf)
	if err != nil {
		t.Errorf("initServer() failed: %v", err)
	}
}

func TestHandler(t *testing.T) {
	template := template.Must(template.New("test").Parse(onSearchPayload))
	handler := mockHandler(template)

	tests := []struct {
		payload    []byte
		wantStatus int
	}{
		{
			payload:    testSearchRequest,
			wantStatus: http.StatusOK,
		},
		{
			payload:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			payload:    testSearchRequestUncomplete,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		request := httptest.NewRequest(http.MethodPost, "/search", bytes.NewReader(test.payload))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if got, want := response.Code, test.wantStatus; got != want {
			t.Errorf("Handler got status %d, want %d", got, want)
		}
	}
}
