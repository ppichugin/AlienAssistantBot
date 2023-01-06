# "Alien Assistant" - Telegram Bot

---
### The functionality:

1. This Telegram-bot receives currency-pair (like USD/EUR) and provides their current exchange rate. 
2. Secrets keeper ... (in process)

---

#### Application uses:
* Go v.1.19
* [go-telegram-bot-api-library](https://github.com/go-telegram-bot-api/telegram-bot-api)
* [Viper](https://github.com/spf13/viper)
* Docker


#### API for currency updates:
* [Exchange Rates Data API](https://apilayer.com/marketplace/exchangerates_data-api)


> **_NOTE:_** 'config.env' file is not included! It contains two environment variables:
    ```
    TELEGRAM_APITOKEN=... & APILayerKey=...
    ```

---

### Usage:

1. Add Telegram-bot by the following link: https://t.me/AlienAssistantBot

2. Pull docker image contained its backend: 
```bash
docker pull petrodev/alien-assistant-bot:latest
```

3. Run docker image:
```bash
docker run -it -d petrodev/alien-assistant-bot:latest
```

4. Possible commands:
   * /start - simply start talking to bot
   * /menu - Main menu
   * /help - Help message
   * /rate - Send currency pair in format `USD/EUR` and get the current exchange rate
   * /sayhi - Get nice "Hi" :)
   * to be continued..


5. Enjoy with **Alien Assistant bot**!