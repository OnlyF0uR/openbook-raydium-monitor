package discord_hook

import (
	"context"
	"fmt"
	"os"

	"github.com/OnlyF0uR/solana-monitor/internal/hooks"
	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/gagliardetto/solana-go"
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

func tokenToSymbol(token solana.PublicKey) string {
	if token == solana.WrappedSol {
		return "SOL"
	} else if token == utils.USDC_MINT_PUBKEY {
		return "USDC"
	} else {
		return "N/A"
	}
}

func tokenHelper(ctx context.Context, token solana.PublicKey) (*utils.TokenData, *utils.TokenMeta) {
	btd, err := utils.GetTokendata(ctx, token, false)
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			if err.Error() == "failed to get mint account data" {
				color.New(color.FgYellow).Printf("Token (%s) is not a valid mint\n", token.String())
				return nil, nil
			}
			color.New(color.FgYellow).Printf("Error getting token data (%s): %v\n", token.String(), err)
		}
		return nil, nil
	}

	if btd.Data.Uri == "" {
		color.New(color.FgYellow).Printf("Token (%s) data had no metadata URI, skipping it.\n", btd.Mint.String())
		return nil, nil
	}

	btm, err := utils.FetchTokenMeta(btd.Data.Uri)
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			color.New(color.FgYellow).Printf("Error fetching token meta (URI: %s): %v\n", btd.Data.Uri, err)
		}
		return nil, nil
	}

	if btm.Description == "" {
		btm.Description = "None"
	}

	if len(btm.Description) > 600 {
		btm.Description = btm.Description[:600] + "..."
	}

	return btd, btm
}
