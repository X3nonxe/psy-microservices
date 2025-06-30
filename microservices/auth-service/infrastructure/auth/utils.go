package auth

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

// Generate UUID v4
func GenerateUUID() string {
	return uuid.NewString()
}

// Generate random salt
func generateRandomSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Encode Argon2 hash dengan format standar
func encodeArgon2Hash(hash, salt []byte) string {
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	return "$argon2id$v=19$m=65536,t=1,p=4$" + b64Salt + "$" + b64Hash
}
