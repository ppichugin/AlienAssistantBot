package secretkeeper

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

// get retrieves secret from DB and decrypts it.
func get(args []string, update *tgbotapi.Update) error {

	var passphrase string

	db := config.GlobConf.Database
	chatID := update.Message.Chat.ID
	owner := update.SentFrom().UserName

	// Parse the command arguments
	if len(args) < 2 {
		utils.SendMessage(chatID, "invalid command. Use /get <title>")
		return ErrInvalidCmd
	}

	title := args[1]

	// TODO: check if more than one secret found with the same title

	// Retrieve the secret from the database
	row := db.QueryRow(
		"SELECT id, title, message, passphrase, iv, expiration, reads_remaining, owner FROM secrets WHERE title=$1 AND owner=$2",
		title,
		owner,
	)
	secret := Secret{}

	err := row.Scan(
		&secret.ID,
		&secret.Title,
		&secret.Message,
		&secret.Passphrase,
		&secret.IV,
		&secret.Expiration,
		&secret.ReadsRemaining,
		&secret.Owner,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: secret not found", err)
	}
	if err != nil { //nolint:wsl
		return fmt.Errorf("%w: error retriving from DB", err)
	}

	// Delete secret if it's expired
	if secret.Expiration.Before(time.Now()) {
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			return fmt.Errorf("%w: error deleting secret", err)
		}

		return fmt.Errorf("%w: secret has expired and deleted", err)
	}

	// Obtaining passphrase from the user to encrypt secret
	utils.SendMessage(chatID, "Enter passphrase to decrypt the secret:")

	for upd := range *config.GlobConf.BotUpdatesCh {
		if upd.Message == nil {
			continue
		}
		passphrase = upd.Message.Text //nolint:wsl

		break
	}

	// Generate encryption key based on the given passphrase
	encryptionKey, err := GenCryptoKey(passphrase)
	if err != nil {
		return fmt.Errorf("%w: error generating enryption key", err)
	}

	// decrypt the secret
	err = secret.decrypt(encryptionKey, passphrase)
	if err != nil {
		return fmt.Errorf("%w: error decrypting", err)
	}

	// Check if the secret is for a single read only
	if secret.ReadsRemaining == 1 {
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			return fmt.Errorf("%w: error deleting from DB", err)
		}

		utils.SendMessage(chatID, "This is the last read. Secret deleted.")
	}

	// Decrement the number of reads remaining
	if secret.ReadsRemaining > 1 {
		_, err := db.Exec("UPDATE secrets SET reads_remaining = reads_remaining - 1 WHERE id=$1", secret.ID)
		if err != nil {
			return fmt.Errorf("%w: error updating reads_remaining in DB", err)
		}

		secret.ReadsRemaining--
		utils.SendMessage(chatID, "Reads Remaining decreased.")
	}

	// Send the secret back to the user
	utils.SendMessage(chatID, secret.String())
	secret = Secret{} //nolint:wsl

	return nil
}
