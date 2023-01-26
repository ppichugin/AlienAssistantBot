package secretkeeper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

// GenCryptoKey generates a cryptographic key from the user's passphrase.
// The key that is generated from this function is of 32 bytes (AES-256 bits).
func GenCryptoKey(passphrase string) ([]byte, error) {
	cost := 1 << 14 //nolint:gomnd    // Controls the CPU/memory cost (power of two: 2^14)
	rounds := 8     // Number of rounds
	par := 1        // Number of goroutines
	keyLen := 32    // Length of the derived key, in bytes

	key, err := scrypt.Key([]byte(passphrase), nil, cost, rounds, par, keyLen)
	if err != nil {
		return nil, fmt.Errorf("error generating CryptoKey: %w", err)
	}

	return key, nil
}

// encrypt encrypts the secret using AES in CTR mode.
func (s *Secret) encrypt(key []byte) error {
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
	passphraseBytes := []byte(s.Passphrase)
	messageBytes := []byte(s.Message)

	// Add padding.
	passphraseBytes = pad(passphraseBytes)
	messageBytes = pad(messageBytes)

	// encrypt the passphrase.
	stream.XORKeyStream(passphraseBytes, passphraseBytes)
	s.Passphrase = base64.StdEncoding.EncodeToString(passphraseBytes)

	// encrypt the message.
	stream.XORKeyStream(messageBytes, messageBytes)
	s.Message = base64.StdEncoding.EncodeToString(messageBytes)

	return nil
}

// decrypt decrypts the secret using AES in CTR mode.
func (s *Secret) decrypt(key []byte, passphrase string) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	stream := cipher.NewCTR(block, s.IV)
	passphraseBytes, err := decryptPartial(s.Passphrase, stream)
	if err != nil {
		return err
	}

	if string(passphraseBytes) != passphrase {
		return ErrPassphrase
	}

	messageBytes, err := decryptPartial(s.Message, stream)
	if err != nil {
		return err
	}

	s.Passphrase = string(passphraseBytes)
	s.Message = string(messageBytes)

	return nil
}

func decryptPartial(strToDecrypt string, stream cipher.Stream) ([]byte, error) {
	decodeString, err := base64.StdEncoding.DecodeString(strToDecrypt)
	if err != nil {
		return nil, err
	}

	stream.XORKeyStream(decodeString, decodeString)
	decodeString = unpad(decodeString)

	return decodeString, nil
}

// pad adds padding to input slice.
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
