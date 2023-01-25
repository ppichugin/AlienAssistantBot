package secretkeeper

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

// Get retrieves secret from DB and decrypts it
func Get(args []string, update *tgbotapi.Update) error {
	bot := config.GlobConf.BotAPIConfig
	db := config.GlobConf.Database
	chatID := update.Message.Chat.ID

	// Parse the command arguments
	if len(args) < 2 {
		text := "invalid command. Use /get <name>"
		utils.SendMessage(chatID, text)
		return fmt.Errorf(text)
	}
	name := args[1]
	owner := update.SentFrom().UserName

	// TODO: check if more than one secret found with the same name

	// Retrieve the secret from the database
	row := db.QueryRow(
		"SELECT id, name, username, password, iv, expiration, reads_remaining, owner FROM secrets WHERE name=$1 AND owner=$2",
		name,
		owner,
	)
	secret := Secret{}
	err := row.Scan(
		&secret.ID,
		&secret.Name,
		&secret.Username,
		&secret.Password,
		&secret.IV,
		&secret.Expiration,
		&secret.ReadsRemaining,
		&secret.Owner,
	)

	// TODO: should stay in this mode, do not return

	if errors.Is(err, sql.ErrNoRows) {
		text := "secret not found"
		utils.SendMessage(chatID, text)
		return fmt.Errorf("%s: no rows in db (%w)", text, err)
	}
	if err != nil {
		text := "error retrieving secret"
		utils.SendMessage(chatID, text)
		return fmt.Errorf("%s: something went wrong in DB (%w)", text, err)
	}

	// Check if the secret has expired, and if so - delete it.
	if secret.Expiration.Before(time.Now()) {
		text := "secret has expired"
		utils.SendMessage(chatID, text)
		if _, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID); err != nil {
			text := "Error deleting secret"
			utils.SendMessage(chatID, text)
			return fmt.Errorf("%s: something went wrong in DB (%w)", text, err)
		}
		return fmt.Errorf("%s: and deleted from DB (%w)", text, err)
	}

	// Check if the secret is for a single read only
	if secret.ReadsRemaining == 1 {
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			text := "Error deleting secret"
			utils.SendMessage(chatID, text)
			return fmt.Errorf("%s: something went wrong in DB (%w)", text, err)
		}
	}

	// Decrement the number of reads remaining
	if secret.ReadsRemaining > 1 {
		_, err := db.Exec("UPDATE secrets SET reads_remaining = reads_remaining - 1 WHERE id=$1", secret.ID)
		secret.ReadsRemaining--
		utils.SendMessage(chatID, "Reads Remaining decreased and now equal to: "+strconv.Itoa(secret.ReadsRemaining))
		if err != nil {
			text := "Error updating ReadsRemaining"
			utils.SendMessage(chatID, text)
			return fmt.Errorf("%s: something went wrong in DB (%w)", text, err)
		}
	}

	utils.SendMessage(chatID, "Enter passphrase to decrypt the secret:")
	var passphrase string
	for upd := range *config.GlobConf.BotUpdatesCh {
		if upd.Message == nil {
			continue
		}
		passphrase = upd.Message.Text
		break
	}

	// Generate encryption encryptionKey based on the given passphrase
	encryptionKey, err := GenCryptoKey(passphrase, bot)
	if err != nil {
		log.Println(err)
		return err
	}

	// Decrypt the secret
	if err := secret.Decrypt(encryptionKey, passphrase); err != nil {
		text := "Error decrypting secret"
		utils.SendMessage(chatID, text)
		return fmt.Errorf("%s: something went wrong Decrypt (%w)", text, err)
	}

	// TODO add structure with String() method for decrypted secrets?

	// Send the secret back to the user
	response := fmt.Sprintf("Username: %s\nPassword: %s", secret.Username, secret.Password)
	if secret.Owner != "" {
		response += fmt.Sprintf("\nOwner: %s", secret.Owner)
	}
	utils.SendMessage(chatID, response)
	secret = Secret{}

	return nil
}
