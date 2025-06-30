package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	keyLength    = 32
)

// Argon2ID Hash dengan salt acak
func Argon2Hash(password string) (string, error) {
	salt, err := generateRandomSalt(16)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		keyLength,
	)

	return encodeArgon2Hash(hash, salt), nil
}

// Verifikasi password
func Argon2Verify(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		keyLength,
	)

	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1
}
