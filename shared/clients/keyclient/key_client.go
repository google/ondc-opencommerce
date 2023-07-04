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

// Package keyclient provides functionalities to access all kinds of keys.
package keyclient

import (
	"context"
	"encoding/json"
	"fmt"

	sm "cloud.google.com/go/secretmanager/apiv1"
	smpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	smrpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
)

// SecretManagerKeyClient provides keys for encryption and authentication.
type SecretManagerKeyClient struct {
	secretClient *sm.Client
	projectID    string
	secretID     string
}

// New create a new SecretManagerKeyService.
func New(ctx context.Context, projectID, secretID string, opts ...option.ClientOption) (*SecretManagerKeyClient, error) {
	secretClient, err := sm.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP Secret Manager client: %v", err)
	}

	client := &SecretManagerKeyClient{
		secretClient: secretClient,
		projectID:    projectID,
		secretID:     secretID,
	}
	return client, nil
}

// Close closes the underlying connection.
func (c *SecretManagerKeyClient) Close() error {
	return c.secretClient.Close()
}

// ServiceSigningPrivateKeyset provides the ED25519 private key of our service.
func (c *SecretManagerKeyClient) ServiceSigningPrivateKeyset(ctx context.Context) ([]byte, error) {
	result, err := c.readSecret(ctx)
	if err != nil {
		return nil, err
	}
	return result["signingKey"]["signingKeySet"], nil
}

// ServiceEncryptionPrivateKey provides the X25519 private key of our service.
func (c *SecretManagerKeyClient) ServiceEncryptionPrivateKey(ctx context.Context) ([]byte, error) {
	result, err := c.readSecret(ctx)
	if err != nil {
		return nil, err
	}
	return result["encryptionKey"]["privateKeyEncryption"], nil
}

// AddKey adds a new secret version to a given secret ID with a given payload.
func (c *SecretManagerKeyClient) AddKey(ctx context.Context, secretID string, payload []byte) error {
	request := &smpb.AddSecretVersionRequest{
		Parent: secretID,
		Payload: &smrpb.SecretPayload{
			Data: payload,
		},
	}
	_, err := c.secretClient.AddSecretVersion(ctx, request)
	return err
}

func (c *SecretManagerKeyClient) readSecret(ctx context.Context) (map[string]map[string][]byte, error) {
	req := &smpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", c.projectID, c.secretID),
	}
	keyData, err := c.secretClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}

	var result map[string]map[string][]byte
	err = json.Unmarshal(keyData.Payload.Data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
