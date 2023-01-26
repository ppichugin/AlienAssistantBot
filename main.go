package main

import (
	"fmt"
	"log"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/services/exchangerates"
	"github.com/ppichugin/AlienAssistantBot/services/secretkeeper"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

func main() {

	v := viper.New()
	v.AddConfigPath("config")

	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("fatal error config file: %w", err))
	}

	v.AutomaticEnv()

	config.GlobConf = config.Configuration{
		TelegramAPIToken:   v.GetString("TELEGRAM_APITOKEN"),
		ExchangeRateAPIKey: v.GetString("APILayerKey"),
		HostDB:             v.GetString("hostDB"),
		PortDB:             v.GetString("portDB"),
		UserDB:             v.GetString("userDB"),
		PasswordDB:         v.GetString("passwordDB"),
		NameDB:             v.GetString("dbname"),
	}

	bot, err := tgBotApi.NewBotAPI(config.GlobConf.TelegramAPIToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgBotApi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)
	config.GlobConf.BotUpdatesCh = &updates
	config.GlobConf.BotAPIConfig = bot

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgBotApi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = config.StartMsg
			case "help":
				msg.Text = config.HelpMsg
			case "sayhi":
				msg.Text = fmt.Sprintf("Hi, %s 😃!", update.SentFrom().FirstName)
			case "status":
				msg.Text = "I'm ok."
			case "rate":
				exchangerates.StartRate(&update)
				msg.Text = config.HelpMsg
			case "secret":
				secretkeeper.StartSecret(&update)
				msg.Text = config.HelpMsg
			case "menu":
				msg.Text = config.HelpMsg
			default:
				msg.Text = config.ErrMsg
			}

			utils.SendMessage(update.Message.Chat.ID, msg.Text)
		} else {
			utils.SendMessage(update.Message.Chat.ID, config.IncorrectCmdFormat)
		}
	}
}
