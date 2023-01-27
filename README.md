# "Alien Assistant" - Telegram Bot

[![Go Report Card](https://goreportcard.com/badge/github.com/ppichugin/AlienAssistantBot)](https://goreportcard.com/report/github.com/ppichugin/AlienAssistantBot)

---

### The functionality:

1. This Telegram-bot receives currency-pair (like USD/EUR) and provides their current exchange rate.
2. Secrets keeper (in develop):
   1. save your secret encrypted with your own passphrase
   2. obtain a secret by name and decrypt it with your passphrase
   3. you can specify the period to keep the secret
   4. you can specify the amount of reads for the secret
   5. transfer the secret to another Telegram user

---

#### Application uses:

* Go v.1.19
* [go-telegram-bot-api-library](https://github.com/go-telegram-bot-api/telegram-bot-api)
* [Viper](https://github.com/spf13/viper)
* Docker
* Docker-compose

#### API for currency updates:

* [Exchange Rates Data API](https://apilayer.com/marketplace/exchangerates_data-api)

> **_NOTE:_** 'config.env' file is not included! It contains two environment variables:
```
TELEGRAM_APITOKEN=... & APILayerKey=...
```

---

### Usage:

1. Add Telegram-bot by the following link: https://t.me/AlienAssistantBot

2. Run from root directory of the repository:

   ```bash
   make run-pull
   ```
   This will run docker-compose that will pull the latest images with backend application
   
   > **_NOTE:_** Since 'config.env' file is not included into the repo, you can not re-build backend app yourself.

3. Available commands:
    * `/start` - simply start talking to bot
    * `/menu` - Main menu
    * `/help` - Help message
    * `/sayhi` - Get nice "Hi" :)
    * `/rate` - Send currency pair in format `USD/EUR` and get the current exchange rate
    * `/secret` - go to Secrets Keeper mode:
      * `/save <name> <username> <passphrase> [expiration] [reads] [owner]`
        > **_`[expiration]` option:_** P{x}Y{x}M{x}DT{x}M{x}S. Example: "`P12Y4MT15M`" where T is a separator for Time (minutes, seconds)
        
        > **_Other options:_** `reads` = natural number, `owner` = Telegram username
      * `/get <name>`
        * enter passphrase to decrypt
      * `/transfer` ... to be continued

4. Enjoy with **Alien Assistant bot**!