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

// Package registryclient provides a service to communicate with the ONDC registry.
package registryclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	log "github.com/golang/glog"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/registry"
)

// RegistryClient communicates with the ONDC registry.
type RegistryClient struct {
	httpClient *http.Client
	baseURL    *url.URL

	lookupURL    string
	subscribeURL string

	ondcEnvironment string
}

// New create a new RegistryClient.
func New(registryURL, ondcEnvironment string) (*RegistryClient, error) {
	baseURL, err := url.Parse(registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry URL: %v", err)
	}

	lookupURL := baseURL.JoinPath("lookup").String()
	subscribeURL := baseURL.JoinPath("subscribe").String()

	return &RegistryClient{
		httpClient:      &http.Client{},
		baseURL:         baseURL,
		lookupURL:       lookupURL,
		subscribeURL:    subscribeURL,
		ondcEnvironment: ondcEnvironment,
	}, nil
}

// PublicSigningKey looks up a signing public key (ED25519) from the ONDC registry.
func (c *RegistryClient) PublicSigningKey(subscriberID, uniqueKeyID string, ondcCtx model.Context) ([]byte, error) {
	requestBody := registry.LookupRequest{
		SubscriberID: &subscriberID,
		UkID:         uniqueKeyID,
		// Country:      ondcCtx.Country,
		// Domain:       &ondcCtx.Domain.Value,
		// City:         ondcCtx.City,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	if c.ondcEnvironment == "staging" {
		// rename the key in JSON
		requestBodyJSON = bytes.Replace(requestBodyJSON, []byte("ukId"), []byte("unique_key_id"), 1)
	}

	response, err := c.httpClient.Post(c.lookupURL, "application/json", bytes.NewReader(requestBodyJSON))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBodyRaw, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusOK {
		log.Infof("Lookup keys: body %s", responseBodyRaw)
		log.Infof("Lookup keys: status code %d", response.StatusCode)
		return nil, errors.New("Error fetching a signing public key")
	}

	var responseBody registry.LookupResponse
	if err := json.Unmarshal(responseBodyRaw, &responseBody); err != nil {
		return nil, err
	}
	if len(responseBody) == 0 {
		log.Errorf("DEBUG: lookup_request: %v", requestBody)
		log.Errorf("DEBUG: lookup_request_json: %q", requestBodyJSON)
		return nil, errors.New("Public Signing Keys are not found")
	}

	return base64.StdEncoding.DecodeString(responseBody[0].SigningPublicKey)
}

// RotateKeys do the keys rotation via Registry /subscribe API.
func (c *RegistryClient) RotateKeys(encryptionPublicKey, signingPublicKey, requestID, subscriberID string, rotationPeriod time.Duration) error {
	currentTime := time.Now().UTC()
	validUntil := currentTime.Add(rotationPeriod)
	requestBody := registry.SubscribeRequest{
		Context: &registry.SubscribeContext{
			Operation: &registry.Context{OpsNo: 6}, // Buyer/Non-MSN/MSN SellerApp key rotation
		},
		Message: &registry.SubscribeMessage{
			RequestID: requestID,
			Timestamp: registry.CustomTime(currentTime),
			Entity: &registry.Entity{
				SubscriberID: subscriberID,
				KeyPair: &registry.KeyPair{
					SigningPublicKey:    signingPublicKey,
					EncryptionPublicKey: encryptionPublicKey,
					ValidFrom:           registry.CustomTime(currentTime),
					ValidUntil:          registry.CustomTime(validUntil),
				},
			},
		},
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	log.Infof("rotatekeys: Request JSON: %q", requestJSON)

	request, err := http.NewRequest(http.MethodPost, c.subscribeURL, bytes.NewReader(requestJSON))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var subscribeRes registry.SubscribeResponse
	if err := json.Unmarshal(responseBody, &subscribeRes); err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK || subscribeRes.Message.Ack.Status != "ACK" {
		return fmt.Errorf("Key rotation error: %v", subscribeRes.Error)
	}
	return nil
}
