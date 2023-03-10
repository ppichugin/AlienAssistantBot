package secretkeeper

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

// save encrypts secret and saves it to DB.
func save(args []string, update *tgbotapi.Update) error {
	db := config.GlobConf.Database
	chatID := update.Message.Chat.ID

	if len(args) < 4 {
		text := "Invalid command. Use /save <name> <username> <passphrase> [expiration - P{x}Y{x}M{x}DT{x}M{x}S] [reads] [owner]"
		utils.SendMessage(chatID, text)

		return fmt.Errorf("incorrect syntax: %w", ErrInvalidCmd)
	}

	name := args[1]
	username := args[2]
	passphrase := args[3]

	// If the expiration is not	specified - it will 31-12-3000
	expiration := time.Date(3000, time.December, 31, 23, 59, 0, 0, time.UTC)
	if len(args) > 4 {
		d := utils.Duration(args[4])
		expiration = time.Now().Add(d)
	}

	var reads int
	var err error

	if len(args) > 5 {
		reads, err = strconv.Atoi(args[5])
		if err != nil {
			utils.SendMessage(chatID, err.Error())
			return fmt.Errorf("%w: incorrect duration (%s)", err, args[5])
		}
	}

	// If the owner is not specified, then the current username will be saved as owner
	owner := update.SentFrom().UserName
	if len(args) > 6 {
		owner = args[6]
	}

	// Generate encryption key based on the given passphrase
	encryptionKey, err := GenCryptoKey(passphrase)
	if err != nil {
		log.Println(err)
		utils.SendMessage(chatID, err.Error())

		return err
	}

	secret := Secret{
		Title:          name,
		Message:        username,
		Passphrase:     passphrase,
		Expiration:     expiration,
		ReadsRemaining: reads,
		Owner:          owner,
	}

	if err := secret.encrypt(encryptionKey); err != nil {
		utils.SendMessage(chatID, err.Error())

		return fmt.Errorf("error in encryption module: %w", err)
	}

	// save the secret to the database
	_, err = db.Exec(
		"INSERT INTO secrets (title, message, passphrase, iv, expiration, reads_remaining, owner) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		secret.Title,
		secret.Message,
		secret.Passphrase,
		secret.IV,
		secret.Expiration,
		secret.ReadsRemaining,
		secret.Owner,
	)
	if err != nil {
		text := "Error saving secret"
		utils.SendMessage(chatID, text)

		return fmt.Errorf("%s: something went wrong in DB (%w)", text, err)
	}

	utils.SendMessage(chatID, "Secret saved successfully")
	secret = Secret{} //nolint:wsl

	return nil
}
