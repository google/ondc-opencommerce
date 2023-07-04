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

// Package registryclienttest provide a stub for registryclient.RegistryClient
package registryclienttest

import (
	"fmt"
	"time"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/models/model"
)

// Stub stubs registryclient.RegistryClient
type Stub struct {
	signingKey []byte
}

// NewStub creates a new stub.
func NewStub() *Stub {
	return &Stub{}
}

// RotateKeys does nothing and return no error.
func (*Stub) RotateKeys(string, string, string, string, time.Duration) error {
	return nil
}

// PublicSigningKey does nothing and return no error.
func (s *Stub) PublicSigningKey(string, string, model.Context) ([]byte, error) {
	if s.signingKey == nil {
		return nil, fmt.Errorf("registry cleint stub: no public signing key")
	}
	return s.signingKey, nil
}

// SetKey sets the stored signing key.
func (s *Stub) SetKey(signingKey []byte) {
	s.signingKey = signingKey
}
