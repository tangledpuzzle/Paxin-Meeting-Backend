package initializers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ConnectTelegram(config *Config) (*tgbotapi.BotAPI, error) {
	
	bot, err := tgbotapi.NewBotAPI(config.TELEGRAM_TOKEN)
	if err != nil {
		return nil, err
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return bot, nil

}