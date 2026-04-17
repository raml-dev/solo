// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/denisbrodbeck/machineid"
)

// getEncryptionKey derives a 32-byte key from the unique MachineID using SHA-256.
func getEncryptionKey() ([]byte, error) {
	id, err := machineid.ID()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(id))
	return hash[:], nil
}

// Encrypt encrypts the given plaintext using AES-256-GCM and the MachineID-derived key.
// It returns a slice containing the nonce followed by the ciphertext.
func Encrypt(plaintext []byte) ([]byte, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the ciphertext to the nonce and returns the result.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts the given ciphertext (which should include the nonce)
// using AES-256-GCM and the MachineID-derived key.
func Decrypt(data []byte) ([]byte, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
