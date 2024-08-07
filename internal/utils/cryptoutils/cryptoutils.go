package cryptoutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
)

// RandomBytes - returns a random byte array of the given size.
// Randomness is crypto safe.
func RandomBytes(size int) []byte {
	result := make([]byte, size)
	exc.Must(exc.ReturnAnyErr(rand.Read(result)))

	return result
}

// CreateGCM - creates a new GCM cipher with the given key.
// The key must be 32 characters long because we use AES-256.
func CreateGCM(key string) cipher.AEAD {
	if len(key) != 32 {
		panic("key length must be 32 characters long")
	}
	block := exc.MustResult(aes.NewCipher([]byte(key)))
	gcm := exc.MustResult(cipher.NewGCM(block))

	return gcm
}

// Encrypt - encrypts the given plaintext with the given passphrase.
func Encrypt(plaintext, passphrase string) string {
	gcm := CreateGCM(passphrase)
	nonce := RandomBytes(gcm.NonceSize())

	return base64.StdEncoding.EncodeToString(gcm.Seal(
		nonce,
		nonce,
		[]byte(plaintext),
		nil,
	))
}

// Decrypt - decrypts the given ciphertext with the given passphrase.
func Decrypt(cipherText, key string) string {
	gcm := CreateGCM(key)

	nonceSize := gcm.NonceSize()

	ciphertextByte := exc.MustResult(base64.StdEncoding.DecodeString(cipherText))
	nonce, ciphertextByteClean := ciphertextByte[:nonceSize], ciphertextByte[nonceSize:]

	return string(exc.MustResult(gcm.Open(
		nil,
		nonce,
		ciphertextByteClean,
		nil,
	)))
}
