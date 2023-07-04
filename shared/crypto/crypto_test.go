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

package crypto

import (
	"testing"

	"partner-innovation.googlesource.com/googleondcaccelerator.git/shared/cryptotest"
)

func TestEncryptDecryptFromExample(t *testing.T) {
	example := cryptotest.NewExample(t)

	// Compare the results with examples
	encryptedText, err := EncryptMessage(example.PlainText, example.X25519PrivateKey, example.X25519PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if encryptedText != example.EncryptedText {
		t.Errorf("got %q, want %q", encryptedText, example.EncryptedText)
	}

	decryptedText, err := DecryptMessage(example.EncryptedText, example.X25519PrivateKey, example.X25519PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if decryptedText != example.PlainText {
		t.Errorf("got %q, want %q", decryptedText, example.PlainText)
	}
}

func TestEncryptDecryptMessage(t *testing.T) {
	ondcPrivateKey, ondcPublicKey, _, err := GenerateEncryptionKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	npPrivateKey, npPublicKey, _, err := GenerateEncryptionKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	challengeString := "This is a secret message"

	// On the ONDC Registry's side
	encryptedMessage, err := EncryptMessage(challengeString, ondcPrivateKey, npPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	// On the Network Participant's side
	t.Logf("Encrypted message: %q", encryptedMessage)
	decryptedMessage, err := DecryptMessage(encryptedMessage, npPrivateKey, ondcPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	if decryptedMessage != challengeString {
		t.Errorf("got %q, want %q", decryptedMessage, challengeString)
	}
}
