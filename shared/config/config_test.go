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

package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const testConfigDir = "testdata/"

func TestReadOnboardingConfigSuccess(t *testing.T) {
	const filename = "onboarding.json"
	filepath := (testConfigDir + filename)
	want := OnboardingConfig{
		ProjectID:             "bit-ondc",
		Port:                  8080,
		RequestID:             "340dff12-661c-40b1-8aba-e4c20046aeda",
		SecretID:              "test-secret",
		RegistryEncryptPubKey: "MCowBQYDK2VuAyEAa9Wbpvd9SsrpOZFcynyt/TO3x0Yrqyys4NUGIvyxX2Q=",
	}

	got, err := Read[OnboardingConfig](filepath)
	if err != nil {
		t.Fatalf("ReadConfig(%q) failed unexpectedly; err=%v", filename, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadConfig(%q) mismatch (-want +got):\n%s", filename, diff)
	}
}

func TestReadBPPAPIConfigSuccess(t *testing.T) {
	const filename = "bpp_api.json"
	filepath := (testConfigDir + filename)
	want := BPPAPIConfig{
		SubscriberID: "bpp.com",
		ProjectID:    "test-project",
		TopicID:      "test-topic",
		Port:         8080,
		RegistryURL:  "https://preprod.registry.ondc.org/ondc",
		GatewayURL:   "https://preprod.gateway.ondc.org",
		InstanceID:   "test-instance",
		DatabaseID:   "test-database",
	}

	got, err := Read[BPPAPIConfig](filepath)
	if err != nil {
		t.Fatalf("ReadConfig(%q) failed unexpectedly; err=%v", filename, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadConfig(%q) mismatch (-want +got):\n%s", filename, diff)
	}
}

func TestReadCallbackActionConfigSuccess(t *testing.T) {
	const filename = "callback_action.json"
	filepath := (testConfigDir + filename)
	want := CallbackActionConfig{
		ProjectID:      "test-project",
		SecretID:       "test-secret",
		TopicID:        "test-topic",
		SubscriptionID: []string{"test-subscription"},
		InstanceID:     "test-instance",
		DatabaseID:     "test-database",
		GatewayURL:     "https://preprod.gateway.ondc.org",
		SubscriberID:   "bpp.com",
		SubscriberURL:  "https://bpp.com/api",
		KeyID:          "test-key",
	}

	got, err := Read[CallbackActionConfig](filepath)
	if err != nil {
		t.Fatalf("ReadConfig(%q) failed unexpectedly; err=%v", filename, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadConfig(%q) mismatch (-want +got):\n%s", filename, diff)
	}
}

func TestReadConfigFailed(t *testing.T) {
	filenames := []string{
		"non_exist.json",
		"invalid.json",
		"invalid_key_rotation.json",
	}

	for _, filename := range filenames {
		filepath := (testConfigDir + filename)

		_, err := Read[BuyerAppConfig](filepath)

		if err == nil { // If NO error
			t.Errorf("ReadConfig(%q) succeeded unexpectedly", filename)
		}
	}
}
