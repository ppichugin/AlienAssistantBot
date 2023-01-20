package secretkeeper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/crypto/scrypt"
)

// GenCryptoKey generates a cryptographic key from the user's passphrase
// The key that is generated from this function is of 32 bytes (AES-256 bits)
func GenCryptoKey(password string, m *tgbotapi.Message, bot *tgbotapi.BotAPI) ([]byte, error) {
	key, err := scrypt.Key([]byte(password), nil, 1<<14, 8, 1, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Unable to generate encryption key on the passphrase. Call admin")
		bot.Send(msg)

		// For tests TODO remove
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))

		log.Printf(fmt.Sprintf("Error: %s", err))
		return nil, fmt.Errorf(msg.Text + ": " + err.Error())
	}
	return key, nil
}

// Encrypt encrypts the secret using AES in CTR mode
func (s *Secret) Encrypt(key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Generate a random initialization vector
	s.IV = make([]byte, aes.BlockSize)
	if _, err := rand.Read(s.IV); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, s.IV)
	passwordBytes := []byte(s.Password)

	// Add padding
	passwordBytes = pad(passwordBytes)

	// Encrypt the password
	stream.XORKeyStream(passwordBytes, passwordBytes)
	s.Password = base64.StdEncoding.EncodeToString(passwordBytes)

	return nil
}

// Decrypt decrypts the secret using AES in CTR mode
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

	// Decrypt the password
	stream.XORKeyStream(passwordBytes, passwordBytes)
	// Removes padding
	passwordBytes = unpad(passwordBytes)

	if string(passwordBytes) != passphrase {
		return errors.New("invalid passphrase")
	}
	s.Password = string(passwordBytes)

	return nil
}

// pad add padding to input slice
func pad(src []byte) []byte {
	padLen := aes.BlockSize - len(src)%aes.BlockSize
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(src, pad...)
}

// unpad removes padding from input slice
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
