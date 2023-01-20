package secretkeeper

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

// Get retrieves secret from DB and decrypts it
func Get(args []string, bot *tgbotapi.BotAPI, update *tgbotapi.Update, db *sql.DB) error {
	// Parse the command arguments
	m := update.Message
	if len(args) < 2 {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Invalid command. Use /get <name>")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}

	name := args[1]
	owner := update.SentFrom().UserName

	// TODO: check if more then one secret found with the same name
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
	//todo - remove after checking
	log.Println("", secret)

	if err == sql.ErrNoRows {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Secret not found")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}
	if err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Error retrieving secret")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}

	// Check if the secret has expired
	if secret.Expiration.Before(time.Now()) {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Secret has expired")
		bot.Send(msg)
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Error deleting secret")
			bot.Send(msg)
			return fmt.Errorf(msg.Text)
		}
		return fmt.Errorf(msg.Text)
	}

	// Check if the secret is for a single read only
	if secret.ReadsRemaining == 1 {
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Error deleting secret")
			bot.Send(msg)
			return fmt.Errorf(msg.Text)
		}
	} else if secret.ReadsRemaining > 1 {
		// Decrement the number of reads remaining
		_, err := db.Exec("UPDATE secrets SET reads_remaining = reads_remaining - 1 WHERE id=$1", secret.ID)
		if err != nil {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Error updating secret.ReadsRemaining")
			bot.Send(msg)
			return fmt.Errorf(msg.Text)
		}
	}

	utils.SendMessage(update, bot, fmt.Sprintf("Enter passphrase to decrypt"))
	var passphrase string
	for upd := range *config.GlobConf.BotUpdatesCh {
		if upd.Message == nil { // ignore any non-Message updates
			continue
		}
		passphrase = upd.Message.Text
		break
	}

	// Generate encryption key based on the given passphrase
	key, err := GenCryptoKey(passphrase, m, bot)
	if err != nil {
		// For tests TODO remove
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error in GenCryptoKey: %s", err)))
		return err
	}

	// Decrypt the secret
	if err := secret.Decrypt(key, passphrase); err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Error decrypting secret")
		bot.Send(msg)
		// For tests TODO remove
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))
		return fmt.Errorf(msg.Text)
	}

	// Send the secret back to the user
	response := fmt.Sprintf("Username: %s\nPassword: %s", secret.Username, secret.Password)
	if secret.Owner != "" {
		response += fmt.Sprintf("\nOwner: %s", secret.Owner)
	}
	msg := tgbotapi.NewMessage(m.Chat.ID, response)
	bot.Send(msg)
	return nil
}
