package main

import (
	"fmt"
	"log"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/services"
)

func main() {

	x := viper.New()
	x.AddConfigPath("config")
	err := x.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	x.AutomaticEnv()
	config.GlobConf = config.Configuration{
		TelegramAPIToken:   x.GetString("TELEGRAM_APITOKEN"),
		ExchangeRateAPIKey: x.GetString("APILayerKey"),
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

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgBotApi.NewMessage(update.Message.Chat.ID, "")
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
				services.GetRates(&update, bot, &updates)
				msg.Text = config.HelpMsg
			case "menu":
				msg.Text = config.HelpMsg
			default:
				msg.Text = config.ErrMsg
			}
			bot.Send(msg)
		} else {
			msg.Text = "Please use command style starts with '/'. For example, /status"
			bot.Send(msg)
		}
	}
}
