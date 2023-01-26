package utils

import (
	"log"
	"regexp"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
)

func SendMessage(chatID int64, text string) {
	bot := config.GlobConf.BotAPIConfig
	msg := tgbotapi.NewMessage(chatID, text)

	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func IsValidPair(s string) bool {
	re := regexp.MustCompile("^[a-zA-Z]+/[a-zA-Z]+$") // `EUR/USD`

	return re.MatchString(s)
}

func SplitArgs(s string) []string {
	return strings.Fields(strings.TrimSpace(s))
}
