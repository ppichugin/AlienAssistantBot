package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/scrypt"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/utils"

	_ "github.com/lib/pq"
)

// Secret represents a secret saved in the database
type Secret struct {
	ID             uuid.UUID
	Name           string
	Username       string
	Password       string
	IV             []byte
	Expiration     time.Time
	ReadsRemaining int
	Owner          string
}

func SecretKeeper(update *tgBotApi.Update, bot *tgBotApi.BotAPI) {

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
		case upd := <-*config.GlobConf.BotUpdatesCh:
			cmd = upd.Message.Text
			if !upd.Message.IsCommand() {
				utils.SendMessage(update, bot, config.IncorrectCmdFormat)
				utils.SendMessage(update, bot, "Please repeat")
				break
			}
			args = append(args, strings.Split(cmd, " ")...)

			key := args[0]

			cmd = key[1:]

			switch cmd {
			case "save":
				err := Save(args, bot, &upd, db)
				if err != nil {
					log.Println("Error saving: ", err)
				}
			case "get":
				err := Get(args, bot, &upd, db)
				if err != nil {
					log.Println("Error getting secret: ", err)
				}
			case "menu":
				return
			}
			args = make([]string, 0, 7)

		}
	}

}

// Save encrypts secret and saves it to DB
func Save(args []string, bot *tgBotApi.BotAPI, update *tgBotApi.Update, db *sql.DB) error {
	m := update.Message
	if len(args) < 4 {
		msg := tgBotApi.NewMessage(m.Chat.ID,
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
			msg := tgBotApi.NewMessage(m.Chat.ID, "Error parsing reads")
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
	key, err := genCryptoKey(password, m, bot)
	if err != nil {
		// For tests TODO remove
		bot.Send(tgBotApi.NewMessage(m.Chat.ID, fmt.Sprintf("Error in genCryptoKey: %s", err)))
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

// genCryptoKey generates a cryptographic key from the user's passphrase
// The key that is generated from this function is of 32 bytes (AES-256 bits)
func genCryptoKey(password string, m *tgBotApi.Message, bot *tgBotApi.BotAPI) ([]byte, error) {

	key, err := scrypt.Key([]byte(password), nil, 1<<14, 8, 1, 32)
	if err != nil {
		//TODO: handle error & change to util method SendMessage
		msg := tgBotApi.NewMessage(m.Chat.ID, "Unable to generate encryption key on the passphrase. Call admin")
		bot.Send(msg)

		// For tests TODO remove
		bot.Send(tgBotApi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))

		log.Printf(fmt.Sprintf("Error: %s", err))
		return nil, fmt.Errorf(msg.Text + ": " + err.Error())
	}
	return key, nil
}

// Get retrieves secret from DB and decrypts it
func Get(args []string, bot *tgBotApi.BotAPI, update *tgBotApi.Update, db *sql.DB) error {
	// Parse the command arguments
	m := update.Message
	if len(args) < 2 {
		msg := tgBotApi.NewMessage(m.Chat.ID, "Invalid command. Use /get <name>")
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
		msg := tgBotApi.NewMessage(m.Chat.ID, "Secret not found")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}
	if err != nil {
		msg := tgBotApi.NewMessage(m.Chat.ID, "Error retrieving secret")
		bot.Send(msg)
		return fmt.Errorf(msg.Text)
	}

	// Check if the secret has expired
	if secret.Expiration.Before(time.Now()) {
		msg := tgBotApi.NewMessage(m.Chat.ID, "Secret has expired")
		bot.Send(msg)
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			msg := tgBotApi.NewMessage(m.Chat.ID, "Error deleting secret")
			bot.Send(msg)
			return fmt.Errorf(msg.Text)
		}
		return fmt.Errorf(msg.Text)
	}

	// Check if the secret is for a single read only
	if secret.ReadsRemaining == 1 {
		_, err := db.Exec("DELETE FROM secrets WHERE id=$1", secret.ID)
		if err != nil {
			msg := tgBotApi.NewMessage(m.Chat.ID, "Error deleting secret")
			bot.Send(msg)
			return fmt.Errorf(msg.Text)
		}
	} else if secret.ReadsRemaining > 1 {
		// Decrement the number of reads remaining
		_, err := db.Exec("UPDATE secrets SET reads_remaining = reads_remaining - 1 WHERE id=$1", secret.ID)
		if err != nil {
			msg := tgBotApi.NewMessage(m.Chat.ID, "Error updating secret.ReadsRemaining")
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
	key, err := genCryptoKey(passphrase, m, bot)
	if err != nil {
		// For tests TODO remove
		bot.Send(tgBotApi.NewMessage(m.Chat.ID, fmt.Sprintf("Error in genCryptoKey: %s", err)))
		return err
	}

	// Decrypt the secret
	if err := secret.Decrypt(key, passphrase); err != nil {
		msg := tgBotApi.NewMessage(m.Chat.ID, "Error decrypting secret")
		bot.Send(msg)
		// For tests TODO remove
		bot.Send(tgBotApi.NewMessage(m.Chat.ID, fmt.Sprintf("Error: %s", err)))
		return fmt.Errorf(msg.Text)
	}

	// Send the secret back to the user
	response := fmt.Sprintf("Username: %s\nPassword: %s", secret.Username, secret.Password)
	if secret.Owner != "" {
		response += fmt.Sprintf("\nOwner: %s", secret.Owner)
	}
	msg := tgBotApi.NewMessage(m.Chat.ID, response)
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
