package config

const (
	RatesAPIUrl = "https://api.apilayer.com/exchangerates_data/latest"

	StartMsg = "Hi there! I'm an Alien Assistant bot. I can help you with exchange rates and keep your secrets"
	HelpMsg  = "Menu: \n" +
		"/menu 		- Main menu\n" +
		"/status 	- Check status,\n" +
		"/sayhi 	- Nice welcoming message :)\n" +
		"/rate 		- Exchange rates on the currency pair like 'USD/EUR'"
	ErrMsg     = "I'm sorry, I didn't understand that command. Use /help to see a list of available commands."
	GetRateMsg = "Please enter currency pair in format 'USD/EUR'"
)

type Configuration struct {
	TelegramAPIToken   string
	ExchangeRateAPIKey string
}

var GlobConf Configuration
