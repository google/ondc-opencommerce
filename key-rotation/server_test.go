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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/keyclienttest"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/clients/registryclienttest"
)

func TestRotationHandler(t *testing.T) {
	keyClient := keyclienttest.NewStub(t)
	registryClient := registryclienttest.NewStub()
	server := initServer(keyClient, registryClient, config{})

	tests := []struct {
		name               string
		requestMethod      string
		requestBody        string
		responseStatusCode int
		responseBody       string
	}{
		{
			name:               "Invalid HTTP Method",
			requestMethod:      http.MethodGet,
			requestBody:        "",
			responseStatusCode: http.StatusMethodNotAllowed,
			responseBody:       "",
		},
		{
			name:               "Event Type Is Not 'SECRET_ROTATE'",
			requestMethod:      http.MethodPost,
			requestBody:        `{"message": {"attributes": {"dataFormat": "JSON_API_V1", "eventType": "SECRET_CREATE", "secretId": "projects/1019248048664/secrets/key-logistic", "timestamp": "2023-03-23T00:05:00.122233-07:00"}, "data":"eyJuYW1lIjoicHJvamVjdHMvMTAxOTI0ODA0ODY2NC9zZWNyZXRzL2tleXMiLCJyZXBsaWNhdGlvbiI6eyJhdXRvbWF0aWMiOnt9fSwiY3JlYXRlVGltZSI6IjIwMjMtMDMtMjNUMDY6Mzk6MjkuMTczODY0WiIsInRvcGljcyI6W3sibmFtZSI6InByb2plY3RzL29uZGMtYnV5ZXItZGV2L3RvcGljcy9rZXktcm90YXRpb24tYjZjZTJjYjYifV0sImV0YWciOiJcIjE1Zjc4YmU1MWY5Yzc5XCIiLCJyb3RhdGlvbiI6eyJuZXh0Um90YXRpb25UaW1lIjoiMjAyMy0wOS0yMVQyMjoyNTowMFoiLCJyb3RhdGlvblBlcmlvZCI6IjE1NzgwMDAwcyJ9fQ==", "messageId":"7231722999366000", "message_id": "7231722999366000", "publishTime":"2023-03-23T07:05:01.854Z", "publish_time": "2023-03-23T07:05:01.854Z"}, "subscription": ""}`,
			responseStatusCode: http.StatusOK,
			responseBody:       `Ignore event type: "SECRET_CREATE"`,
		},
		{
			name:               "Valid Request",
			requestMethod:      http.MethodPost,
			requestBody:        `{"message": {"attributes": {"dataFormat": "JSON_API_V1", "eventType": "SECRET_ROTATE", "secretId": "projects/1019248048664/secrets/key-logistic", "timestamp": "2023-03-23T00:05:00.122233-07:00"}, "data":"eyJuYW1lIjoicHJvamVjdHMvMTAxOTI0ODA0ODY2NC9zZWNyZXRzL2tleXMiLCJyZXBsaWNhdGlvbiI6eyJhdXRvbWF0aWMiOnt9fSwiY3JlYXRlVGltZSI6IjIwMjMtMDMtMjNUMDY6Mzk6MjkuMTczODY0WiIsInRvcGljcyI6W3sibmFtZSI6InByb2plY3RzL29uZGMtYnV5ZXItZGV2L3RvcGljcy9rZXktcm90YXRpb24tYjZjZTJjYjYifV0sImV0YWciOiJcIjE1Zjc4YmU1MWY5Yzc5XCIiLCJyb3RhdGlvbiI6eyJuZXh0Um90YXRpb25UaW1lIjoiMjAyMy0wOS0yMVQyMjoyNTowMFoiLCJyb3RhdGlvblBlcmlvZCI6IjE1NzgwMDAwcyJ9fQ==", "messageId":"7231722999366000", "message_id": "7231722999366000", "publishTime":"2023-03-23T07:05:01.854Z", "publish_time": "2023-03-23T07:05:01.854Z"}, "subscription": ""}`,
			responseStatusCode: http.StatusOK,
			responseBody:       "Key rotation is completed",
		},
	}

	for _, test := range tests {
		request := httptest.NewRequest(test.requestMethod, "/", strings.NewReader(test.requestBody))
		response := httptest.NewRecorder()

		server.rotationHandler(response, request)

		if got, want := response.Code, test.responseStatusCode; got != want {
			t.Errorf("Name: %s ==> Status Code got %v, want %v", test.name, got, want)
		}
		if got, want := response.Body.String(), test.responseBody; got != want {
			t.Errorf("Name: %s ==> Body got %v, want %v", test.name, got, want)
		}
	}
}
