package discord_hook

import (
	"fmt"
	"os"

	"github.com/OnlyF0uR/solana-monitor/internal/hooks"
	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

var raydiumChannelID string
var openbookChannelID string

func Initialise() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		panic("DISCORD_BOT_TOKEN not set")
	}

	dc, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord = dc

	raydiumChannelID = os.Getenv("DISCORD_RAYDIUM_CHANNEL")
	if raydiumChannelID == "" {
		panic("DISCORD_RAYDIUM_CHANNEL not set")
	}

	openbookChannelID = os.Getenv("DISCORD_OPENBOOK_CHANNEL")
	if openbookChannelID == "" {
		panic("DISCORD_OPENBOOK_CHANNEL not set")
	}

	// Setup hooks
	hooks.RegisterOpenbookHook(dc_openbook_hook)
	hooks.RegisterRaydiumHook(dc_raydium_hook)

	fmt.Printf("Discord hook initialised\n")
}
