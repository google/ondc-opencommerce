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

// Package cryptotest provides helpers for writing unit tests which involve encryption of
// the subscription flow.
package cryptotest

import (
	"crypto/ecdh"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

// These example values are specified in this ONDC implementation reference.
// https://github.com/ONDC-Official/reference-implementations/tree/main/utilities/signing_and_verification
const (
	examplePrivateKeyDERB64 = "MC4CAQAwBQYDK2VuBCIEIOgl3rf3arbk1PvIe0C9TZp7ImR71NSQdvuSu+zzY6xo"
	ExamplePublicKeyDERB64  = "MCowBQYDK2VuAyEAi801MjVpgFOXHjliyT6Nb14HkS5dj1p41qbeyU6/SC8="
	examplePlainText        = "ONDC is a Great Initiative!"
	exampleEncryptedTextB64 = "CrwN248HS4CIYsUvxtrK0pWCBaoyZh4LnWtGqeH7Mpc="
)

// Example is a set of values to test encryption of the subscription flow.
type Example struct {
	X25519PrivateKey []byte // Private key of one side
	X25519PublicKey  []byte // Public key of another side
	PlainText        string // Plain challenge string
	EncryptedText    string // Encrypted challenge string in base64
}

// NewExample creates a new instance of Example.
//
// Every instance of Example has the same values.
func NewExample(t *testing.T) *Example {
	t.Helper()

	// Extract the raw public key
	publicKeyDER, err := base64.StdEncoding.DecodeString(ExamplePublicKeyDERB64)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	publicKeyParsed, err := x509.ParsePKIXPublicKey(publicKeyDER)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	publicKeyCasted, ok := publicKeyParsed.(*ecdh.PublicKey)
	if !ok {
		t.Fatal("setup failed: incorrect key type")
	}
	publicKeyRaw := publicKeyCasted.Bytes()
	if len(publicKeyRaw) != 32 {
		t.Fatalf("setup failed: invalid key length")
	}

	// Extract the raw private key
	privateKeyDER, err := base64.StdEncoding.DecodeString(examplePrivateKeyDERB64)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	privateKeyParsed, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	privateKeyCasted, ok := privateKeyParsed.(*ecdh.PrivateKey)
	if !ok {
		t.Fatal("setup failed: incorrect key type")
	}
	privateKeyRaw := privateKeyCasted.Bytes()
	if len(privateKeyRaw) != 32 {
		t.Fatalf("setup failed: invalid key length")
	}

	return &Example{
		X25519PrivateKey: privateKeyRaw,
		X25519PublicKey:  publicKeyRaw,
		PlainText:        examplePlainText,
		EncryptedText:    exampleEncryptedTextB64,
	}
}
