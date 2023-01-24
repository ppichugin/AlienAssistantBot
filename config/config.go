package config

import (
	"database/sql"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	RatesAPIUrl = "https://api.apilayer.com/exchangerates_data/latest"

	StartMsg = "Hi there! I'm an Alien Assistant bot. I can help you with current exchange rates and I can safely hide your secrets."
	HelpMsg  = "Menu: \n" +
		"/menu 		- Main menu\n" +
		"/status 	- Check status,\n" +
		"/sayhi 	- Nice welcoming message :)\n" +
		"/rate 		- Exchange rates on the currency pair like 'USD/EUR'\n" +
		"/secret    - Secrets Keeper"
	ErrMsg        = "I'm sorry, I didn't understand that command. Use /help to see a list of available commands."
	GetRateMsg    = "Please enter currency pair in format 'USD/EUR'"
	KeeperHelpMsg = "Supported commands and their formats\n" +
		"Save secret:\n" +
		"/save <name> <username> <password> [expiration] [reads] [owner]\n" +
		"Retrieve saved secret:\n" +
		"/get - ...\n" +
		"/menu - go back to the main menu\n"
	IncorrectCmdFormat = "Please use command style starts with '/'."
)

type Configuration struct {
	TelegramAPIToken   string
	ExchangeRateAPIKey string
	HostDB             string
	PortDB             string
	UserDB             string
	PasswordDB         string
	NameDB             string
	BotUpdatesCh       *tgBotApi.UpdatesChannel
	BotAPIConfig       *tgBotApi.BotAPI
	Database           *sql.DB
}

var GlobConf Configuration
