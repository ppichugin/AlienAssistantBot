package secretkeeper

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/satori/go.uuid"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/model"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

// Errors for the package.
var (
	ErrPassphrase = errors.New("invalid passphrase")
	ErrInvalidCmd = errors.New("invalid command style")
)

// Secret represents a secret saved in the database.
type Secret struct {
	ID             uuid.UUID
	Title          string
	Message        string
	Passphrase     string
	IV             []byte
	Expiration     time.Time
	ReadsRemaining int
	Owner          string
}

// StartSecret starts the dialog for Secrets Keeper mode.
func StartSecret(update *tgBotApi.Update) {
	maxArgs := 7 // max number of arguments including command
	chatID := update.Message.Chat.ID

	err := model.NewDB(3) //nolint:gomnd
	if err != nil {
		utils.SendMessage(chatID, err.Error())
		log.Println(err)
	}

	db := config.GlobConf.Database
	if db == nil {
		return
	}

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	utils.SendMessage(chatID, "You are in Secrets Keeper mode")
	utils.SendMessage(chatID, config.KeeperHelpMsg)

	args := make([]string, 0, maxArgs)

	// Listen commands in secrets mode.
	for upd := range *config.GlobConf.BotUpdatesCh {
		if update.Message == nil {
			continue
		}

		if !upd.Message.IsCommand() {
			utils.SendMessage(chatID, config.IncorrectCmdFormat)
			utils.SendMessage(chatID, "Please repeat")

			continue
		}

		if update.Message.IsCommand() {
			args = append(args, utils.SplitArgs(upd.Message.Text)...)
			key := args[0]
			cmd := key[1:]

			switch cmd {
			case "save":
				err := save(args, &upd)
				if err != nil {
					utils.SendMessage(chatID, err.Error())
					log.Println("Error saving secret: ", err)
				}
			case "get":
				err := get(args, &upd)
				if err != nil {
					utils.SendMessage(chatID, err.Error())
					log.Println("Error getting secret: ", err)
				}
			case "menu":
				return
			}

			args = make([]string, 0, maxArgs)
		}
	}
}

func (s *Secret) String() string {
	return fmt.Sprintf(
		"Title: %s\n"+
			"Message: %s\n"+
			"Passphrase: %s\n"+
			"Expiration: %s\n"+
			"Reads remaining: %d",
		s.Title,
		s.Message,
		s.Passphrase,
		s.Expiration.Format(time.RFC822Z),
		s.ReadsRemaining,
	)
}
