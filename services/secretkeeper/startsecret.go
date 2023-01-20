package secretkeeper

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/satori/go.uuid"

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

func StartSecret(update *tgBotApi.Update) {
	bot := config.GlobConf.BotAPIConfig
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
