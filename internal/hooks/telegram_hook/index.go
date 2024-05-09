package telegram_hook

import (
	"fmt"
	"os"

	"github.com/OnlyF0uR/solana-monitor/internal/hooks"
	"github.com/go-telegram/bot"
)

var telegram *bot.Bot
var chatId string

func Initialise() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		panic("TELEGRAM_BOT_TOKEN not set")
	}

	chatId = os.Getenv("TELEGRAM_CHAT_ID")
	if chatId == "" {
		panic("TELEGRAM_CHAT_ID not set")
	}

	b, err := bot.New(botToken)
	if err != nil {
		panic(err)
	}

	telegram = b

	hooks.RegisterOpenbookHook(tg_openbook_hook)
	hooks.RegisterRaydiumHook(tg_raydium_hook)

	fmt.Printf("Telegram hook initialised\n")
}
