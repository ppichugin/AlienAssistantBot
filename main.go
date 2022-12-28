package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/models"
)

const (
	baseURL = "https://api.apilayer.com/exchangerates_data/latest"

	startMessage = "Hi there! I'm a currency exchange rate bot. Send me a currency pair (e.g. USD/EUR) and I'll get the current exchange rate for you."
	helpMessage  = "To get the current exchange rate for a currency pair, just send me the pair (e.g. USD/EUR). \n" +
		"I'll do my best to get the most recent rate for you.\n" +
		"Other commands: \n" +
		"/status - check status,\n" +
		"/sayhi - just welcoming message :)"
	errMessage = "I'm sorry, I didn't understand that command. Use /start to get started or /help to see a list of available commands."
)

var globConf config.Configuration

func main() {

	x := viper.New()
	x.AddConfigPath("config")
	err := x.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	x.AutomaticEnv()
	globConf = config.Configuration{
		TelegramAPIToken:   x.GetString("TELEGRAM_APITOKEN"),
		ExchangeRateAPIKey: x.GetString("APILayerKey"),
	}

	bot, err := tgBotApi.NewBotAPI(globConf.TelegramAPIToken)
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

		if update.Message.IsCommand() {
			msg := tgBotApi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "start":
				msg.Text = startMessage
			case "help":
				msg.Text = helpMessage
			case "sayhi":
				msg.Text = "Hi, there! 😃"
			case "status":
				msg.Text = "I'm ok."
			default:
				msg.Text = errMessage
			}
			bot.Send(msg)
		} else {
			currencyPair := update.Message.Text

			exchangeRate, err := getExchangeRate(currencyPair)
			if err != nil {
				msg := tgBotApi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("I'm sorry, I was unable to get the exchange rate for %s. Please make sure you've entered a valid currency pair.", currencyPair))
				bot.Send(msg)
				continue
			}

			msg := tgBotApi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("The current exchange rate for %s is %.2f.", currencyPair, exchangeRate))
			bot.Send(msg)
		}
	}
}

func getExchangeRate(currencyPair string) (float64, error) {
	url := fmt.Sprintf("%s?base=%s&symbols=%s", baseURL, currencyPair[0:3], currencyPair[4:7])

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", globConf.ExchangeRateAPIKey)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	var exchangeRates models.ExchangeRates
	err = json.NewDecoder(resp.Body).Decode(&exchangeRates)
	if err != nil {
		return 0, err
	}

	rate, ok := exchangeRates.Rates[currencyPair[4:7]]
	if !ok {
		return 0, fmt.Errorf("invalid currency pair")
	}

	return rate, nil
}
