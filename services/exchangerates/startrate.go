package exchangerates

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ppichugin/AlienAssistantBot/config"
	"github.com/ppichugin/AlienAssistantBot/model"
	"github.com/ppichugin/AlienAssistantBot/utils"
)

// StartRate represents service to provide the current exchange rate for currency pairs.
func StartRate(update *tgBotApi.Update) {
	chatID := update.Message.Chat.ID
	utils.SendMessage(chatID, config.GetRateMsg)

	var currencyPair string

	for upd := range *config.GlobConf.BotUpdatesCh {
		chatID := upd.Message.Chat.ID
		if upd.Message.Command() == "menu" {
			utils.SendMessage(chatID, "Going to main menu.")

			return
		}
		
		currencyPair = upd.Message.Text
		if !utils.IsValidPair(currencyPair) {
			utils.SendMessage(chatID, "Incorrect currency pair. Please repeat in format 'XXX/YYY'")

			continue
		}

		utils.SendMessage(chatID, fmt.Sprintf("%s, here is the exchange rate you requested: ", upd.SentFrom().FirstName))

		break
	}

	exchangeRate, err := getExchangeRate(currencyPair)
	if err != nil {
		log.Println(err)
		utils.SendMessage(chatID, fmt.Sprintf("Error: %v", err))

		return
	}

	utils.SendMessage(chatID, fmt.Sprintf(" %s is %.2f.", currencyPair, exchangeRate))
	utils.SendMessage(chatID, "Going to main menu.")
}

func getExchangeRate(currencyPair string) (float64, error) {
	url := fmt.Sprintf("%s?base=%s&symbols=%s",
		config.RatesAPIUrl, currencyPair[0:3], currencyPair[4:7])

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("apikey", config.GlobConf.ExchangeRateAPIKey)

	resp, err := client.Do(req) //nolint:bodyclose
	if err != nil {
		return 0, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var exchangeRates model.ExchangeRates

	err = json.NewDecoder(resp.Body).Decode(&exchangeRates)
	if err != nil {
		return 0, err
	}

	rate, ok := exchangeRates.Rates[currencyPair[4:7]]
	if !ok {
		return 0, fmt.Errorf("invalid currency pair (%s)", currencyPair)
	}

	return rate, nil
}
