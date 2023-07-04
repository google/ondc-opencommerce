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

package authentication

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	pb "github.com/google/tink/go/proto/ed25519_go_proto"
	"google.golang.org/protobuf/proto"
)

// Example values from the ONDC document
var (
	ed22519PrivateKey       = "lP3sHA+9gileOkXYJXh4Jg8tK0gEEMbf9yCPnFpbldhrAY+NErqL9WD+Vav7TE5tyVXGXBle9ONZi2W7o144eQ=="
	ed25519PublicKey        = "awGPjRK6i/Vg/lWr+0xObclVxlwZXvTjWYtlu6NeOHk="
	payload                 = []byte(`{"context":{"domain":"nic2004:60212","country":"IND","city":"Kochi","action":"search","core_version":"0.9.1","bap_id":"bap.stayhalo.in","bap_uri":"https://8f9f-49-207-209-131.ngrok.io/protocol/","transaction_id":"e6d9f908-1d26-4ff3-a6d1-3af3d3721054","message_id":"a2fe6d52-9fe4-4d1a-9d0b-dccb8b48522d","timestamp":"2022-01-04T09:17:55.971Z","ttl":"P1M"},"message":{"intent":{"fulfillment":{"start":{"location":{"gps":"10.108768, 76.347517"}},"end":{"location":{"gps":"10.102997, 76.353480"}}}}}}`)
	creationTime      int64 = 1641287875
	expirationTime    int64 = 1641291475
	currentTime       int64 = 1641290475
	wantedSignature         = "cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="
)

func TestSignAndVerify(t *testing.T) {
	newKeysetJSON, err := GenerateKeysetJSON()
	if err != nil {
		t.Fatalf("GenerateKeysetJSON() failed unexpectedly; err=%v", err)
	}

	rawPublicKey, err := ExtractRawPublicKey(newKeysetJSON)
	if err != nil {
		t.Fatalf("ExtractRawPublicKey() failed unexpectedly; err=%v", err)
	}

	signature, err := SignPayload(payload, newKeysetJSON, creationTime, expirationTime)
	if err != nil {
		t.Fatalf("SignPayload() failed unexpectedly; err=%v", err)
	}

	if err := VerifySignature(signature, payload, rawPublicKey, creationTime, expirationTime); err != nil {
		t.Errorf("VerifySignature() failed unexpectedly; err=%v", err)
	}
}

func TestSignPayload(t *testing.T) {
	// Prepare Tink keyset for testing
	keysetJSON := createTestKeysetJSON(t)

	signature, err := SignPayload(payload, []byte(keysetJSON), creationTime, expirationTime)
	if err != nil {
		t.Fatal(err)
	}

	if signature != wantedSignature {
		t.Errorf("SignPayload() = %v, want = %s", signature, wantedSignature)
	}
}

func TestVerifySignature(t *testing.T) {
	publicKeyByte, err := base64.StdEncoding.DecodeString(ed25519PublicKey)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = VerifySignature(wantedSignature, payload, publicKeyByte, creationTime, expirationTime)

	if err != nil {
		t.Errorf("VerifySignature() failed unexpectedly; err=%v", err)
	}
}

func TestCreateAuthSignature(t *testing.T) {
	const (
		want         = `Signature keyId="example-bap.com|bap1234|ed25519",algorithm="ed25519",created="1641287875",expires="1641291475",headers="(created) (expires) digest",signature="cjbhP0PFyrlSCNszJM1F/YmHDVAWsZqJUPzojnE/7TJU3fJ/rmIlgaUHEr5E0/2PIyf0tpSnWtT6cyNNlpmoAQ=="`
		subscriberID = "example-bap.com"
		keyID        = "bap1234"
	)
	// Prepare Tink keyset for testing
	keysetJSON := createTestKeysetJSON(t)

	got, err := CreateAuthSignature(payload, []byte(keysetJSON), creationTime, expirationTime, subscriberID, keyID)
	if err != nil {
		t.Fatalf("CreateAuthSignature() failed unexpectedly; err=%v", err)
	}
	if got != want {
		t.Errorf("CreateAuthSignature() = %v, want = %s", got, want)
	}
}

// createTestKeysetJSON creates ED25519 keyset from key values in the ONDC document.
func createTestKeysetJSON(t *testing.T) string {
	t.Helper()

	privateKeyRaw, err := base64.StdEncoding.DecodeString(ed22519PrivateKey)
	if err != nil {
		t.Fatalf("Setup: Failed to create test keyset JSON: %s", err)
	}

	publicKeyRaw, err := base64.StdEncoding.DecodeString(ed25519PublicKey)
	if err != nil {
		t.Fatalf("Setup: Failed to create test keyset JSON: %s", err)
	}

	if bytes.Compare(privateKeyRaw[32:], publicKeyRaw) != 0 {
		t.Fatal("Setup: Failed to create test keyset JSON: Invalid public and private keys")
	}

	privateKey := pb.Ed25519PrivateKey{
		Version:  0,
		KeyValue: privateKeyRaw[:32],
		PublicKey: &pb.Ed25519PublicKey{
			Version:  0,
			KeyValue: publicKeyRaw,
		},
	}
	privateKeyOnWire, err := proto.Marshal(&privateKey)
	if err != nil {
		t.Fatalf("Setup: Failed to create test keyset JSON: %s", err)
	}

	privateKeyOnWireB64 := base64.StdEncoding.EncodeToString(privateKeyOnWire)
	return fmt.Sprintf(`{"primaryKeyId":3051498072,"key":[{"keyData":{"typeUrl":"type.googleapis.com/google.crypto.tink.Ed25519PrivateKey","value":"%s","keyMaterialType":"ASYMMETRIC_PRIVATE"},"status":"ENABLED","keyId":3051498072,"outputPrefixType":"RAW"}]}`, privateKeyOnWireB64)
}
