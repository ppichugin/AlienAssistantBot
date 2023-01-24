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

// Save encrypts secret and saves it to DB
func Save(args []string, update *tgbotapi.Update) error {
	bot := config.GlobConf.BotAPIConfig
	db := config.GlobConf.Database
	chatID := update.Message.Chat.ID
	if len(args) < 4 {
		text := "Invalid command. Use /save <name> <username> <passphrase> [expiration - P{x}Y{x}M{x}DT{x}M{x}S] [reads] [owner]"
		utils.SendMessage(chatID, text)
		return fmt.Errorf("incorrect syntax: %s", text)
	}

	name := args[1]
	username := args[2]
	passphrase := args[3]

	// TODO: remove any double/triple spaces between args

	expiration := time.Date(3000, time.December, 31, 23, 59, 0, 0, time.UTC)
	if len(args) > 4 {
		d := utils.Duration(args[4])
		expiration = time.Now().Add(d)
	}

	var reads int
	if len(args) > 5 {
		n, err := strconv.Atoi(args[5])
		if err != nil {
			text := "Error parsing reads"
			utils.SendMessage(chatID, text)
			return fmt.Errorf("%s: something wrong in duration (%w)", text, err)
		}
		reads = n
	}

	// If the owner is not specified, then the current user will be saved
	var owner string
	if len(args) > 6 {
		owner = args[6]
	}
	if owner == "" {
		owner = update.SentFrom().UserName
	}

	// Generate encryption key based on the given passphrase
	encryptionKey, err := GenCryptoKey(passphrase, bot)
	if err != nil {
		log.Println(err)
		utils.SendMessage(chatID, err.Error())
		return err
	}

	secret := Secret{
		Name:           name,
		Username:       username,
		Password:       passphrase,
		Expiration:     expiration,
		ReadsRemaining: reads,
		Owner:          owner,
	}

	if err := secret.Encrypt(encryptionKey); err != nil {
		utils.SendMessage(chatID, err.Error())
		return fmt.Errorf("error in encryption module: %w", err)
	}

	// Save the secret to the database
	_, err = db.Exec(
		"INSERT INTO secrets (name, username, password, iv, expiration, reads_remaining, owner) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		secret.Name,
		secret.Username,
		secret.Password,
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
	secret = Secret{}

	return nil
}
