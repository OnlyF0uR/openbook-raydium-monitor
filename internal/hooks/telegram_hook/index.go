package telegram_hook

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/OnlyF0uR/solana-monitor/internal/hooks"
	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/gagliardetto/solana-go"
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

func getHolderString(ctx context.Context, mint solana.PublicKey, poolCoinTokenAccount solana.PublicKey, supply float64, liquidity float64) string {
	topHolders := utils.GetTopHolders_S(ctx, mint)

	var topHolderRaydiumAmount string
	var topHoldersStr string
	if topHolders == nil {
		topHoldersStr = "N/A"
	} else {
		for i, holder := range *topHolders {
			supplyPct := (holder.Amount / supply) * 100

			var address = "[" + bot.EscapeMarkdown(holder.PublicKey.Short(6)) + "](https://solscan.io/account/" + holder.PublicKey.String() + ")"
			if holder.PublicKey == poolCoinTokenAccount {
				topHolderRaydiumAmount = "*Raydium: " + strings.Replace(strconv.FormatFloat(supplyPct, 'f', 2, 64), ".", "\\.", 1) + "%*\n\n"
				address = address + " \\(LP\\)"
			}
			if i < 5 {
				// We add to string here too
				topHoldersStr += address + " \\- " + strings.Replace(strconv.FormatFloat(supplyPct, 'f', 2, 64), ".", "\\.", 1) + "%\n"
			}
		}
	}
	if len(topHoldersStr) > 0 {
		topHoldersStr = topHoldersStr[:len(topHoldersStr)-1]
	}

	if topHolderRaydiumAmount == "" {
		topHolderRaydiumAmount = "*Raydium: " + strings.Replace(strconv.FormatFloat((liquidity/supply)*100, 'f', 2, 64), ".", "\\.", 1) + " SOL" + "%*\n\n"
	}

	return topHolderRaydiumAmount + topHoldersStr
}
