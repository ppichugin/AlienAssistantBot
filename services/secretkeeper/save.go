package secretkeeper

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/utils"
)

// Save encrypts secret and saves it to DB
func Save(args []string, bot *tgbotapi.BotAPI, update *tgbotapi.Update, db *sql.DB) error {
	m := update.Message
	if len(args) < 4 {
		msg := tgbotapi.NewMessage(m.Chat.ID,
			"Invalid command. Use /save <name> <username> <password> [expiration - P{x}Y{x}M{x}DT{x}M{x}S] [reads] [owner]")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}

	name := args[1]
	username := args[2]
	password := args[3]

	expiration := time.Date(3000, time.December, 31, 23, 59, 0, 0, time.UTC)
	if len(args) > 4 {
		d := utils.Duration(args[4])

		//TODO: if not specified - ?

		//if err != nil {
		//	msg := tgBotApi.NewMessage(m.Chat.ID, "Error parsing expiration")
		//	bot.Send(msg)
		//	return fmt.Errorf(msg.Text + ": " + err.Error())
		//}
		expiration = time.Now().Add(d)
	}

	var reads int
	if len(args) > 5 {
		n, err := strconv.Atoi(args[5])
		if err != nil {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Error parsing reads")
			bot.Send(msg)
			return fmt.Errorf(msg.Text + ": " + err.Error())
		}
		reads = n
	}

	// If the owner is not specified - then it will be a user itself
	var owner string
	if len(args) > 6 {
		owner = args[6]
	}
	if owner == "" {
		owner = update.SentFrom().UserName
	}

	// Generate encryption key based on the given passphrase
	key, err := GenCryptoKey(password, m, bot)
	if err != nil {
		// For tests TODO remove
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error in GenCryptoKey: %s", err)))
		return err
	}

	secret := Secret{
		Name:           name,
		Username:       username,
		Password:       password,
		Expiration:     expiration,
		ReadsRemaining: reads,
		Owner:          owner,
	}

	if err := secret.Encrypt(key); err != nil {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Error encrypting password")
		bot.Send(msg)

		// For tests TODO remove
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))

		return fmt.Errorf(msg.Text + ": " + err.Error())
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
		msg := tgbotapi.NewMessage(m.Chat.ID, "Error saving secret")
		bot.Send(msg)
		return fmt.Errorf(msg.Text + ": " + err.Error())
	}

	msg := tgbotapi.NewMessage(m.Chat.ID, "Secret saved successfully")
	bot.Send(msg)
	return nil
}
