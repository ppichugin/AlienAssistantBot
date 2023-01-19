package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/crypto/scrypt"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/utils"

	_ "github.com/lib/pq"
)

// Secret represents a secret saved in the database
type Secret struct {
	ID             int
	Name           string
	Username       string
	Password       string
	IV             []byte
	Expiration     time.Time
	ReadsRemaining int
	Owner          string
}

func SecretKeeper(update *tgBotApi.Update, bot *tgBotApi.BotAPI, updates *tgBotApi.UpdatesChannel) {

	// Connect to the database
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.GlobConf.HostDB,
		config.GlobConf.PortDB,
		config.GlobConf.UserDB,
		config.GlobConf.PasswordDB,
		config.GlobConf.NameDB)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	utils.SendMessage(update, bot, "You are in Secrets Keeper mode")
	utils.SendMessage(update, bot, config.KeeperHelpMsg)
	var cmd string
	var args = make([]string, 0, 7)

	//outer:
	// Loop to listen commands
	for {
		select {
		case upd := <-*updates:
			cmd = upd.Message.Text
			args = append(args, strings.Split(cmd, " ")...)
			key := args[0]
			if key[:1] != "/" {
				utils.SendMessage(update, bot, config.IncorrectCmdFormat)
				utils.SendMessage(update, bot, "Please repeat")
				continue
			}
			cmd = key[1:]

			switch cmd {
			case "save":
				err := Save(args, bot, &upd, db)
				if err != nil {
					return
				}
			case "menu":
				return
			}

		}
	}

}

// Save saves the encrypted secret to DB
func Save(args []string, bot *tgBotApi.BotAPI, update *tgBotApi.Update, db *sql.DB) error {
	//args := strings.Split(m.Text, " ")
	m := update.Message
	if len(args) < 4 {
		msg := tgBotApi.NewMessage(m.Chat.ID, "Invalid command. Use /save <name> <username> <password> [expiration] [reads] [owner]")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}

	name := args[1]
	username := args[2]
	password := args[3]

	var expiration time.Time
	if len(args) > 4 {
		d, err := time.ParseDuration(args[4])
		if err != nil {
			msg := tgBotApi.NewMessage(m.Chat.ID, "Error parsing expiration")
			bot.Send(msg)
			return fmt.Errorf(msg.Text + ": " + err.Error())
		}
		expiration = time.Now().Add(d)
	}

	var reads int
	if len(args) > 5 {
		n, err := strconv.Atoi(args[5])
		if err != nil {
			msg := tgBotApi.NewMessage(m.Chat.ID, "Error parsing reads")
			bot.Send(msg)
			return fmt.Errorf(msg.Text + ": " + err.Error())
		}
		reads = n
	}

	var owner string
	if len(args) > 6 {
		owner = args[6]
	}

	// Encrypt the secret
	key := []byte("TEST_KEYTEST_KEYTEST_KEYTEST_KEY") // Testing encryption key

	// Generate a cryptographic key from the user's passphrase
	// The key that is generated from this function is of 32 bytes (AES-256 bits)
	key, err := scrypt.Key([]byte(password), nil, 1<<14, 8, 1, 32)
	if err != nil {
		msg := tgBotApi.NewMessage(m.Chat.ID, "Unable to encrypt the password. Call admin")
		bot.Send(msg)

		// For tests TODO remove
		bot.Send(tgBotApi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))

		log.Printf(fmt.Sprintf("Error: %s", err))
		return fmt.Errorf(msg.Text + ": " + err.Error())
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
		msg := tgBotApi.NewMessage(m.Chat.ID, "Error encrypting password")
		bot.Send(msg)

		// For tests TODO remove
		bot.Send(tgBotApi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))

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
		msg := tgBotApi.NewMessage(m.Chat.ID, "Error saving secret")
		bot.Send(msg)
		return fmt.Errorf(msg.Text + ": " + err.Error())
	}

	msg := tgBotApi.NewMessage(m.Chat.ID, "Secret saved successfully")
	bot.Send(msg)
	return nil
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
	passwordBytes = pad(passwordBytes)

	// Encrypt the password
	stream.XORKeyStream(passwordBytes, passwordBytes)
	s.Password = base64.StdEncoding.EncodeToString(passwordBytes)

	return nil
}

// Decrypt decrypts the secret using AES in CTR mode
func (s *Secret) Decrypt(key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	passwordBytes, err := base64.StdEncoding.DecodeString(s.Password)
	if err != nil {
		return err
	}

	stream := cipher.NewCTR(block, s.IV)

	// Decrypt the password
	stream.XORKeyStream(passwordBytes, passwordBytes)
	s.Password = string(passwordBytes)

	return nil
}

// pad add padding to input slice, so it's length is multiple of AES block size
func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}
