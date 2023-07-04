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

// Package config contains config schemas for all services.
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/registry"
)

var validate = model.Validator()

// OnboardingConfig is a config for onboarding service.
type OnboardingConfig struct {
	ProjectID             string `json:"projectID" validate:"required"`
	Port                  int    `json:"port" validate:"required"`
	RequestID             string `json:"requestID" validate:"required"`
	SecretID              string `json:"secretID" validate:"required"`
	RegistryEncryptPubKey string `json:"registryEncryptPubKey" validate:"required"`
	ONDCEnvironment       string `json:"ONDCEnvironment"`
}

// BPPAPIConfig is a config for BPP API service.
type BPPAPIConfig struct {
	SubscriberID    string `json:"subscriberID" validate:"required"`
	ProjectID       string `json:"projectID" validate:"required"`
	TopicID         string `json:"topicID" validate:"required"`
	Port            int    `json:"port" validate:"required"`
	RegistryURL     string `json:"registryURL" validate:"required,url"`
	GatewayURL      string `json:"gatewayURL" validate:"required,url"`
	InstanceID      string `json:"instanceID" validate:"required"`
	DatabaseID      string `json:"databaseID" validate:"required"`
	ONDCEnvironment string `json:"ONDCEnvironment"`
}

// SellerAdapterConfig is a config for seller adapter service.
type SellerAdapterConfig struct {
	ProjectID       string   `json:"projectID" validate:"required"`
	SellerSystemURL string   `json:"sellerSystemURL" validate:"required,url"`
	CallbackTopicID string   `json:"callbackTopicID" validate:"required"`
	SubscriptionID  []string `json:"subscriptionID" validate:"required"`
	ONDCEnvironment string   `json:"ONDCEnvironment"`
}

// CallbackActionConfig is a config for Callback Action Service.
type CallbackActionConfig struct {
	ProjectID      string   `json:"projectID" validate:"required"`
	SecretID       string   `json:"secretID" validate:"required"`
	TopicID        string   `json:"topicID" validate:"required"`
	SubscriptionID []string `json:"subscriptionID" validate:"required"`
	InstanceID     string   `json:"instanceID" validate:"required"`
	DatabaseID     string   `json:"databaseID" validate:"required"`

	// ONDC config
	GatewayURL      string `json:"gatewayURL" validate:"required,url"`
	SubscriberID    string `json:"subscriberID" validate:"required"`
	SubscriberURL   string `json:"subscriberURL" validate:"required,url"`
	KeyID           string `json:"keyID" validate:"required"`
	ONDCEnvironment string `json:"ONDCEnvironment"`
}

// MockRegistryConfig is a config for Mock Registry Service.
type MockRegistryConfig struct {
	Port           int                     `json:"port" validate:"required"`
	Keys           registry.LookupResponse `json:"keys" validate:"required"`
	RegistryKeyset Keyset                  `json:"registryKeyset" validate:"required"`
}

// Keyset is a set of singing/encryption key pairs.
type Keyset struct {
	PublicSigningKey     string `json:"publicSigningKey" validate:"required"`
	PrivateSigningKey    string `json:"privateSigningKey" validate:"required"`
	PublicEncryptionKey  string `json:"publicEncryptionKey" validate:"required"`
	PrivateEncryptionKey string `json:"privateEncryptionKey" validate:"required"`
}

// MockSellerSystemConfig is a config for Mock Seller System.
type MockSellerSystemConfig struct {
	Port int `json:"port" validate:"required"`
}

// MockGatewayConfig is a config for Mock Gateway Service.
type MockGatewayConfig struct {
	Port         int      `json:"port" validate:"required"`
	SubscriberID string   `json:"subscriberID" validate:"required"`
	ProjectID    string   `json:"projectID" validate:"required"`
	SecretID     string   `json:"secretID" validate:"required"`
	RegistryURL  string   `json:"registryURL" validate:"required,url"`
	BPPURLs      []string `json:"bppURLs" validate:"required,dive,url"`
	BAPURLs      []string `json:"bapURLs" validate:"required,dive,url"`

	KeyID           string `json:"keyID" validate:"required"`
	ONDCEnvironment string `json:"ONDCEnvironment"`
}

// BAPAPIConfig is a config for BAP API service.
type BAPAPIConfig struct {
	SubscriberID    string `json:"subscriberID" validate:"required"`
	ProjectID       string `json:"projectID" validate:"required"`
	TopicID         string `json:"topicID" validate:"required"`
	Port            int    `json:"port" validate:"required"`
	RegistryURL     string `json:"registryURL" validate:"required,url"`
	InstanceID      string `json:"instanceID" validate:"required"`
	DatabaseID      string `json:"databaseID" validate:"required"`
	ONDCEnvironment string `json:"ONDCEnvironment"`
}

// RequestActionConfig is a config for Request Action Service.
type RequestActionConfig struct {
	ProjectID      string   `json:"projectID" validate:"required"`
	SubscriptionID []string `json:"subscriptionID" validate:"required"`
	InstanceID     string   `json:"instanceID" validate:"required"`
	DatabaseID     string   `json:"databaseID" validate:"required"`
	SecretID       string   `json:"secretID" validate:"required"`

	// ONDC config
	GatewayURL      string `json:"gatewayURL" validate:"required,url"`
	SubscriberID    string `json:"subscriberID" validate:"required"`
	SubscriberURL   string `json:"subscriberURL" validate:"required,url"`
	KeyID           string `json:"keyID" validate:"required"`
	ONDCEnvironment string `json:"ONDCEnvironment"`
}

// BuyerAppConfig is a config for Buyer App Service.
type BuyerAppConfig struct {
	ProjectID       string `json:"projectID" validate:"required"`
	TopicID         string `json:"topicID" validate:"required"`
	Port            int    `json:"port" validate:"required"`
	ONDCEnvironment string `json:"ONDCEnvironment"`
}

// BuyerAdapterConfig is a config for Buyer Adapter Service.
type BuyerAdapterConfig struct {
	ProjectID       string   `json:"projectID" validate:"required"`
	BuyerAppURL     string   `json:"buyerAppURL" validate:"required,url"`
	SubscriptionID  []string `json:"subscriptionID" validate:"required"`
	ONDCEnvironment string   `json:"ONDCEnvironment"`
}

type config interface {
	OnboardingConfig | BPPAPIConfig | SellerAdapterConfig | CallbackActionConfig |
		MockRegistryConfig | MockSellerSystemConfig | MockGatewayConfig | BAPAPIConfig | RequestActionConfig |
		BuyerAppConfig | BuyerAdapterConfig
}

// Read reads a file from filepath and parses the config file.
func Read[C config](filepath string) (C, error) {
	var config C

	configJSON, err := os.ReadFile(filepath)
	if err != nil {
		return config, fmt.Errorf("cannot read config: %v", err)
	}

	if err := json.Unmarshal(configJSON, &config); err != nil {
		return config, fmt.Errorf("cannot read config: %s", err)
	}

	if err := validate.Struct(config); err != nil {
		return config, fmt.Errorf("cannot read config: %s", err)
	}

	return config, nil
}
