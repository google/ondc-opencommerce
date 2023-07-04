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

// Package keyclienttest provide a stub for keyclient.SecretManagerKeyClient
package keyclienttest

import (
	"context"
	"testing"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/crypto"
	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/signing-authentication/authentication"
)

// Stub stubs keyclient.SecretManagerKeyClient.
type Stub struct {
	serviceSigningPrivateKeyset []byte
	serviceEncryptionPrivateKey []byte
	registryEncryptionPublicKey []byte
}

// NewStub creates a new Stub with newly generated keys.
func NewStub(t *testing.T) *Stub {
	t.Helper()

	keyset, err := authentication.GenerateKeysetJSON()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	privateKey, _, _, err := crypto.GenerateEncryptionKeyPair()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, publicKey, _, err := crypto.GenerateEncryptionKeyPair()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	return &Stub{
		serviceSigningPrivateKeyset: keyset,
		serviceEncryptionPrivateKey: privateKey,
		registryEncryptionPublicKey: publicKey,
	}
}

// NewStubWithKeys creates a new Stub and assigns the given keys if they are not nil.
func NewStubWithKeys(t *testing.T, serviceSigningPrivateKeyset, serviceEncryptionPrivateKey, registryEncryptionPublicKey []byte) *Stub {
	t.Helper()

	stub := NewStub(t)
	if serviceSigningPrivateKeyset != nil {
		stub.serviceSigningPrivateKeyset = serviceSigningPrivateKeyset
	}
	if serviceEncryptionPrivateKey != nil {
		stub.serviceEncryptionPrivateKey = serviceEncryptionPrivateKey
	}
	if registryEncryptionPublicKey != nil {
		stub.registryEncryptionPublicKey = registryEncryptionPublicKey
	}
	return stub
}

// ServiceSigningPrivateKeyset returns the key stored in the stub.
func (s *Stub) ServiceSigningPrivateKeyset(context.Context) ([]byte, error) {
	return s.serviceSigningPrivateKeyset, nil
}

// ServiceEncryptionPrivateKey returns the key stored in the stub.
func (s *Stub) ServiceEncryptionPrivateKey(context.Context) ([]byte, error) {
	return s.serviceEncryptionPrivateKey, nil
}

// RegistryEncryptionPublicKey returns the key stored in the stub.
func (s *Stub) RegistryEncryptionPublicKey(context.Context) ([]byte, error) {
	return s.registryEncryptionPublicKey, nil
}

// AddKey does nothing and return no error.
func (s *Stub) AddKey(context.Context, string, []byte) error { return nil }
