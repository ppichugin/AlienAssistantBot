package secretkeeper

import (
	"log"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/satori/go.uuid"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/model"
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

func StartSecret(update *tgBotApi.Update) {
	chatID := update.Message.Chat.ID

	err := model.NewDB(3)
	if err != nil {
		utils.SendMessage(chatID, err.Error())
		log.Fatal(err)
	}
	db := config.GlobConf.Database
	defer db.Close()

	utils.SendMessage(chatID, "You are in Secrets Keeper mode")
	utils.SendMessage(chatID, config.KeeperHelpMsg)
	var args = make([]string, 0, 7)

	// Listen commands in secrets mode
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
				err := Save(args, &upd)
				if err != nil {
					log.Println("Error saving secret: ", err)
				}
			case "get":
				err := Get(args, &upd)
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
