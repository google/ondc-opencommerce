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

// Package authentication provides functions to sign and verify a payload signature
//
// The implementation details are specified in [Signing Beckn APIs in HTTP].
//
// [Signing Beckn APIs in HTTP]: https://docs.google.com/document/d/1Iw_x-6mtfoMh0KJwL4sqQYM0kD17MLxiMCUOZDBerBo
package authentication

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
	pb "github.com/google/tink/go/proto/ed25519_go_proto"
	"github.com/google/tink/go/signature"
	"github.com/google/tink/go/signature/subtle"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/proto"
)

// GenerateKeysetJSON generates a new ED25519 Tink keyset in JSON format.
func GenerateKeysetJSON() ([]byte, error) {
	keysetHandle, err := keyset.NewHandle(signature.ED25519KeyWithoutPrefixTemplate())
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	jsonWriter := keyset.NewJSONWriter(&buf)
	if err := insecurecleartextkeyset.Write(keysetHandle, jsonWriter); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExtractRawPublicKey extracts a raw ED25519 public key from Tink keyset in JSON format.
func ExtractRawPublicKey(keysetJSON []byte) ([]byte, error) {
	keysetHandle, err := readJSONKeyset(keysetJSON)
	if err != nil {
		return nil, err
	}

	publicKeyHandle, err := keysetHandle.Public()
	if err != nil {
		return nil, err
	}

	var keysetBuffer bytes.Buffer
	jsonWriter := keyset.NewJSONWriter(&keysetBuffer)
	if err := publicKeyHandle.WriteWithNoSecrets(jsonWriter); err != nil {
		return nil, err
	}

	var keysetJSONMap map[string]any
	if err := json.Unmarshal(keysetBuffer.Bytes(), &keysetJSONMap); err != nil {
		return nil, err
	}

	keys, ok := keysetJSONMap["key"].([]any)
	if !ok {
		return nil, fmt.Errorf("Unkown JSON format of Tink keyset")
	}

	key, ok := keys[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Unkown JSON format of Tink keyset")
	}

	keyData, ok := key["keyData"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Unkown JSON format of Tink keyset")
	}

	keyValue, ok := keyData["value"].(string)
	if !ok {
		return nil, fmt.Errorf("Unkown JSON format of Tink keyset")
	}

	tinkKeyValue, err := base64.StdEncoding.DecodeString(keyValue)
	if err != nil {
		return nil, err
	}

	var tinkPublicKey pb.Ed25519PublicKey
	if err := proto.Unmarshal(tinkKeyValue, &tinkPublicKey); err != nil {
		return nil, err
	}

	return tinkPublicKey.GetKeyValue(), nil
}

// Sign creates a signature from a data without constructing a new signing string.
func Sign(data, keysetJSON []byte) ([]byte, error) {
	keyset, err := readJSONKeyset(keysetJSON)
	if err != nil {
		return nil, err
	}

	signer, err := signature.NewSigner(keyset)
	if err != nil {
		return nil, err
	}

	return signer.Sign(data)
}

// SignPayload creates a signature of ONDC payload.
func SignPayload(payload, keysetJSON []byte, createdTimestamp, expiredTimestamp int64) (string, error) {
	signingString := createSigningString(payload, createdTimestamp, expiredTimestamp)
	signature, err := Sign([]byte(signingString), keysetJSON)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies that the signature is well-encoded and signed by the owner of the public key.
func VerifySignature(signature string, payload, publicKey []byte, createdTimestamp, expiredTimestamp int64) error {
	signatureDecoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	verifier, err := subtle.NewED25519Verifier(publicKey)
	if err != nil {
		return err
	}

	signingString := createSigningString(payload, createdTimestamp, expiredTimestamp)
	return verifier.Verify(signatureDecoded, []byte(signingString))
}

// CreateAuthSignature creates a signature string in ONDC format.
func CreateAuthSignature(payload, keysetJSON []byte, createdTimestamp, expiredTimestamp int64, subscriberID, keyID string) (string, error) {
	signature, err := SignPayload(payload, keysetJSON, createdTimestamp, expiredTimestamp)
	if err != nil {
		return "", err
	}

	authSignature := fmt.Sprintf(
		`Signature keyId="%s|%s|ed25519",algorithm="ed25519",created="%d",expires="%d",headers="(created) (expires) digest",signature="%s"`,
		subscriberID,
		keyID,
		createdTimestamp,
		expiredTimestamp,
		signature,
	)
	return authSignature, nil
}

func createSigningString(payload []byte, createdTimestamp, expiredTimestamp int64) string {
	digest := blake2b.Sum512(payload)
	digestB64 := base64.StdEncoding.EncodeToString(digest[:])
	return fmt.Sprintf(
		"(created): %d\n(expires): %d\ndigest: BLAKE-512=%s",
		createdTimestamp,
		expiredTimestamp,
		digestB64,
	)
}

func readJSONKeyset(keysetJSON []byte) (*keyset.Handle, error) {
	keysetReader := bytes.NewReader(keysetJSON)
	jsonReader := keyset.NewJSONReader(keysetReader)
	return insecurecleartextkeyset.Read(jsonReader)
}
