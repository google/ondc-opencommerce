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

// Package crypto provides cryptography-related functions needed for the ONDC subscription flow.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/zenazn/pkcs7pad"
)

var x25519Curve = ecdh.X25519()

// GenerateEncryptionKeyPair generates a new X25519 key pair.
//
// The public key is in DER form.
func GenerateEncryptionKeyPair() (privateKey, publicKey, publicKeyDER []byte, err error) {
	key, err := x25519Curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	publicKeyDER, err = x509.MarshalPKIXPublicKey(key.PublicKey())
	if err != nil {
		return nil, nil, nil, err
	}

	privateKey = key.Bytes()
	publicKey = key.PublicKey().Bytes()
	return privateKey, publicKey, publicKeyDER, nil
}

// DecryptMessage decrypts the encryptedMessage from the ONDC regsitry.
// privateKey and publicKey are X25519 keys. These keys are used to generate a shared secret.
// The shared secret is used as a key of the AES-ECB-PKCS7 algorithm.
func DecryptMessage(encryptedMessage string, privateKey, publicKey []byte) (string, error) {
	aesCipher, err := createAESCipher(privateKey, publicKey)
	if err != nil {
		return "", err
	}

	messageByte, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(messageByte); i += aesCipher.BlockSize() {
		executionSlice := messageByte[i : i+aesCipher.BlockSize()]
		aesCipher.Decrypt(executionSlice, executionSlice)
	}

	messageByte, err = pkcs7pad.Unpad(messageByte)
	if err != nil {
		return "", err
	}

	return string(messageByte), nil
}

// EncryptMessage encrypts a message like how ONDC registry would encrypt it.
// privateKey and publicKey are X25519 keys. These keys are used to generate a shared secret.
// The shared secret is used as a key of the AES-ECB-PKCS7 algorithm.
func EncryptMessage(message string, privateKey, publicKey []byte) (string, error) {
	messageByte := []byte(message)
	aesCipher, err := createAESCipher(privateKey, publicKey)
	if err != nil {
		return "", err
	}

	messageByte = pkcs7pad.Pad(messageByte, aesCipher.BlockSize())

	for i := 0; i < len(messageByte); i += aesCipher.BlockSize() {
		aesCipher.Encrypt(messageByte[i:i+aesCipher.BlockSize()], messageByte[i:i+aesCipher.BlockSize()])
	}

	return base64.StdEncoding.EncodeToString(messageByte), nil
}

func ExtractRawPubKeyFromDER(pubKeyDERB64 string) ([]byte, error) {
	publicKeyDER, err := base64.StdEncoding.DecodeString(pubKeyDERB64)
	if err != nil {
		return nil, fmt.Errorf("crypto: extract raw pub key error: %v", err)
	}

	publicKeyParsed, err := x509.ParsePKIXPublicKey(publicKeyDER)
	if err != nil {
		return nil, fmt.Errorf("crypto: extract raw pub key error: %v", err)
	}

	publicKeyCasted, ok := publicKeyParsed.(*ecdh.PublicKey)
	if !ok {
		return nil, errors.New("crypto: extract raw pub key error: incorrect key type")
	}

	publicKeyRaw := publicKeyCasted.Bytes()
	if keyLenght := len(publicKeyRaw); keyLenght != 32 {
		return nil, fmt.Errorf("crypto: extract raw pub key error: incorrect key type: invalid key length %d", keyLenght)
	}

	return publicKeyRaw, nil
}

func createAESCipher(privateKey, publicKey []byte) (cipher.Block, error) {
	x25519PrivateKey, err := x25519Curve.NewPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	x25519PublicKey, err := x25519Curve.NewPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	sharedSecret, err := x25519PrivateKey.ECDH(x25519PublicKey)
	if err != nil {
		return nil, err
	}

	aesCipher, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, err
	}

	return aesCipher, nil
}
