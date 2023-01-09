package utils

import (
	"regexp"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendMessage(update *tgbotapi.Update, bot *tgbotapi.BotAPI, txt string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
	bot.Send(msg)
}

func IsValidPair(s string) bool {
	re := regexp.MustCompile("^[a-zA-Z]+/[a-zA-Z]+$") // `EUR/USD`
	return re.MatchString(s)
}
