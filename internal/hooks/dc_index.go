package hooks

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/gagliardetto/solana-go"
)

var discord *discordgo.Session

func InitialiseDiscord() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		panic("DISCORD_BOT_TOKEN not set")
	}

	dc, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	discord = dc

	fmt.Printf("Discord hook initialised\n")
}

func tokenHelper(ctx context.Context, baseMint, quoteMint solana.PublicKey) (*utils.TokenData, *utils.TokenData, *utils.TokenMeta) {
	btd, err := utils.GetTokendata(ctx, baseMint, false)
	if err != nil {
		fmt.Printf("Error getting base token data (%s): %v\n", baseMint.String(), err)
		return nil, nil, nil
	}

	qtd, err := utils.GetTokendata(ctx, quoteMint, false)
	if err != nil {
		fmt.Printf("Error getting quote token data (%s): %v\n", quoteMint.String(), err)
		return nil, nil, nil
	}

	if btd.Data.Uri == "" {
		fmt.Printf("Base token (%s) data had to metadata URI\n", btd.Mint.String())
		return nil, nil, nil
	}

	btm, err := utils.FetchTokenMeta(btd.Data.Uri)
	if err != nil {
		fmt.Printf("Error fetching base token meta (URI: %s): %v\n", btd.Data.Uri, err)
		return nil, nil, nil
	}

	if btm.Description == "" {
		btm.Description = "None"
	}

	if len(btm.Description) > 600 {
		btm.Description = btm.Description[:600] + "..."
	}

	return btd, qtd, btm
}

func getRelatedTokenString(ctx context.Context, caller solana.PublicKey, mint string) string {
	relatedTokens, err := utils.GetRelatedTokens(ctx, caller)
	if err != nil {
		return "None"
	}

	formattedRelatedTokens := []string{}
	for _, token := range *relatedTokens {
		if len(formattedRelatedTokens) >= 3 {
			break
		}

		if token.Parsed.Info.Mint == mint {
			continue
		}

		tokenData, err := utils.GetTokendata(ctx, solana.MustPublicKeyFromBase58(token.Parsed.Info.Mint), true)
		if err != nil {
			continue
		}

		formattedRelatedTokens = append(formattedRelatedTokens, "["+tokenData.Data.Symbol+"](https://solscan.io/account/"+token.Parsed.Info.Mint+")")
	}

	relTokensStr := strings.Join(formattedRelatedTokens, ", ")
	if len(relTokensStr) == 0 {
		relTokensStr = "None"
	}

	return relTokensStr
}

func getWarningsString(ctx context.Context, account solana.PublicKey, createdOn string) string {
	warnings := ""

	wsolAmount := utils.GetTokenAmount(ctx, account, solana.WrappedSol)
	if wsolAmount > 0 {
		warnings += "🚨 WSOL detected 🚨\n"
	}

	if createdOn == "https://pump.fun" {
		warnings += "🚨 Pump.fun detected 🚨"
	}

	return warnings
}
