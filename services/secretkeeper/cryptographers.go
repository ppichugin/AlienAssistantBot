package secretkeeper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/crypto/scrypt"
)

// GenCryptoKey generates a cryptographic key from the user's passphrase.
// The key that is generated from this function is of 32 bytes (AES-256 bits).
func GenCryptoKey(password string, bot *tgbotapi.BotAPI) ([]byte, error) {
	cost := 1 << 14 // Controls the CPU/memory cost (power of two: 2^14)
	rounds := 8     // Number of rounds
	par := 1        // Number of goroutines
	keyLen := 32    // Length of the derived key, in bytes

	key, err := scrypt.Key([]byte(password), nil, cost, rounds, par, keyLen)
	if err != nil {
		return nil, fmt.Errorf("error generating CryptoKey: %w", err)
	}

	return key, nil
}

// Encrypt encrypts the secret using AES in CTR mode.
func (s *Secret) Encrypt(key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Generate a random initialization vector.
	s.IV = make([]byte, aes.BlockSize)
	if _, err := rand.Read(s.IV); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, s.IV)
	passwordBytes := []byte(s.Password)

	// Add padding.
	passwordBytes = pad(passwordBytes)

	// Encrypt the password.
	stream.XORKeyStream(passwordBytes, passwordBytes)
	s.Password = base64.StdEncoding.EncodeToString(passwordBytes)

	return nil
}

// Decrypt decrypts the secret using AES in CTR mode.
func (s *Secret) Decrypt(key []byte, passphrase string) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	stream := cipher.NewCTR(block, s.IV)
	passwordBytes, err := base64.StdEncoding.DecodeString(s.Password)
	if err != nil {
		return err
	}

	// Decrypt the password.
	stream.XORKeyStream(passwordBytes, passwordBytes)
	// Removes padding.
	passwordBytes = unpad(passwordBytes)

	if string(passwordBytes) != passphrase {
		return fmt.Errorf("invalid passphrase (%s)", passphrase)
	}
	s.Password = string(passwordBytes)

	return nil
}

// pad add padding to input slice.
func pad(src []byte) []byte {
	padLen := aes.BlockSize - len(src)%aes.BlockSize
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(src, pad...)
}

// unpad removes padding from input slice.
func unpad(src []byte) []byte {
	n := len(src)
	if n == 0 {
		return nil
	}
	padLen := int(src[n-1])
	if padLen > n {
		return nil
	}
	return src[:n-padLen]
}
