package telegram_hook

import (
	"fmt"
	"os"
)

func Initialise() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		panic("TELEGRAM_BOT_TOKEN not set")
	}

	// TODO: intialise bot

	fmt.Printf("Telegram hook initialised\n")
}
