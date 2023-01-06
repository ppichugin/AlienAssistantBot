package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/models"
)

const (
	baseURL = "https://api.apilayer.com/exchangerates_data/latest"

	startMessage = "Hi there! I'm an Alien Assistant bot. I can help you with exchange rates and keep your secrets"
	helpMessage  = "Menu: \n" +
		"/menu 		- Main menu\n" +
		"/status 	- Check status,\n" +
		"/sayhi 	- Nice welcoming message :)\n" +
		"/rate 		- Exchange rates on the currency pair like 'USD/EUR'"
	errMessage = "I'm sorry, I didn't understand that command. Use /help to see a list of available commands."
	getRateMsg = "Please enter currency pair in format 'USD/EUR'"
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

		msg := tgBotApi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = startMessage
			case "help":
				msg.Text = helpMessage
			case "sayhi":
				msg.Text = fmt.Sprintf("Hi, %s 😃!", update.SentFrom().FirstName)
			case "status":
				msg.Text = "I'm ok."
			case "rate":
				getRates(&update, bot, &updates)
				msg.Text = helpMessage
			case "menu":
				msg.Text = helpMessage
			default:
				msg.Text = errMessage
			}
			bot.Send(msg)
		} else {
			msg.Text = "Please use command style starts with '/'. For example, /status"
			bot.Send(msg)
		}
	}
}

func isValidPair(s string) bool {
	re := regexp.MustCompile("^[a-zA-Z]+/[a-zA-Z]+$") // `EUR/USD`
	return re.MatchString(s)
}

func getRates(update *tgBotApi.Update, bot *tgBotApi.BotAPI, updates *tgBotApi.UpdatesChannel) {

	sendMessage(update, bot, getRateMsg)
	var currencyPair string

outer:
	for {
		select {
		case upd := <-*updates:
			currencyPair = upd.Message.Text
			if !isValidPair(currencyPair) {
				sendMessage(&upd, bot, "Incorrect currency pair. Please repeat in format 'XXX/YYY'")
				continue
			}
			sendMessage(&upd, bot, fmt.Sprintf("%s, here we are!\n", upd.SentFrom().FirstName))
			break outer
		}
	}

	exchangeRate, err := getExchangeRate(currencyPair)
	if err != nil {
		sendMessage(update, bot, fmt.Sprintf("Something went wrong. Error: %v", err))
		return
	}
	sendMessage(update, bot, fmt.Sprintf("The current exchange rate for %s is %.2f.", currencyPair, exchangeRate))
}

func sendMessage(update *tgBotApi.Update, bot *tgBotApi.BotAPI, txt string) {
	msg := tgBotApi.NewMessage(update.Message.Chat.ID, txt)
	bot.Send(msg)
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
