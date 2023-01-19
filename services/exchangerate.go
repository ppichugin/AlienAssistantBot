package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/model"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

func GetRates(update *tgBotApi.Update, bot *tgBotApi.BotAPI, updates *tgBotApi.UpdatesChannel) {

	utils.SendMessage(update, bot, config.GetRateMsg)
	var currencyPair string

outer:
	for {
		select {
		case upd := <-*updates:
			if upd.Message.Command() == "menu" {
				utils.SendMessage(update, bot, fmt.Sprintf("Going to main menu."))
				return
			}
			currencyPair = upd.Message.Text
			if !utils.IsValidPair(currencyPair) {
				utils.SendMessage(&upd, bot, "Incorrect currency pair. Please repeat in format 'XXX/YYY'")
				continue
			}
			utils.SendMessage(&upd, bot, fmt.Sprintf("%s, here is the current exchange rate:", upd.SentFrom().FirstName))
			break outer
		}
	}

	exchangeRate, err := getExchangeRate(currencyPair)
	if err != nil {
		utils.SendMessage(update, bot, fmt.Sprintf("Something went wrong. Error: %v", err))
		return
	}
	utils.SendMessage(update, bot, fmt.Sprintf(" %s is %.2f.", currencyPair, exchangeRate))
	utils.SendMessage(update, bot, fmt.Sprintf("Going to main menu."))
}

func getExchangeRate(currencyPair string) (float64, error) {
	url := fmt.Sprintf("%s?base=%s&symbols=%s", config.RatesAPIUrl, currencyPair[0:3], currencyPair[4:7])

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", config.GlobConf.ExchangeRateAPIKey)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	var exchangeRates model.ExchangeRates
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
