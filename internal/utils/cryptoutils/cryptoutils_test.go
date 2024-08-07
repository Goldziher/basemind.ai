package cryptoutils_test

import (
	"github.com/basemind-ai/monorepo/internal/utils/cryptoutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCryptoUtils(t *testing.T) {
	t.Run("RandomBytes", func(t *testing.T) {
		t.Run("should return a random byte array of the given size", func(t *testing.T) {
			assert.Equal(t, 16, len(cryptoutils.RandomBytes(16)))
			assert.Equal(t, 24, len(cryptoutils.RandomBytes(24)))
			assert.Equal(t, 32, len(cryptoutils.RandomBytes(32)))
		})
	})
	t.Run("CreateGCM", func(t *testing.T) {
		t.Run("should create a new GCM cipher with the given key", func(t *testing.T) {
			assert.NotNil(t, cryptoutils.CreateGCM("12345678901234567890123456789012"))
		})
		t.Run("should panic if the key is not 32 characters long", func(t *testing.T) {
			assert.Panics(t, func() {
				cryptoutils.CreateGCM("123456789012345678901234567890")
			})
		})
	})
	t.Run("Encrypt", func(t *testing.T) {
		t.Run("should encrypt the given plaintext with the given passphrase", func(t *testing.T) {
			assert.NotEmpty(t, cryptoutils.Encrypt("test", "12345678901234567890123456789012"))
		})
	})
	t.Run("Decrypt", func(t *testing.T) {
		t.Run("should decrypt the given ciphertext with the given passphrase", func(t *testing.T) {
			plain := "TEST MESSAGE"
			key := string(cryptoutils.RandomBytes(32))
			encrypted := cryptoutils.Encrypt(plain, key)

			assert.Equal(t, plain, cryptoutils.Decrypt(encrypted, key))
		})
	})
}
