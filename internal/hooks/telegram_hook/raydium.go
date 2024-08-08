package telegram_hook

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/raydium"
	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func tg_raydium_hook(msg *raydium.RaydiumInfo, ctx context.Context) {

	baseTokenData, baseTokenMeta := utils.TokenHelper(ctx, msg.BaseMint)
	if baseTokenData == nil || baseTokenMeta == nil {
		return
	}

	tokenBSymbol := utils.TokenToSymbol(msg.QuoteMint)

	// Parse the supply to a float64
	parsedSupply := float64(baseTokenData.Supply) / math.Pow10(int(baseTokenData.Decimals))

	// Get the balance of the caller
	balance := utils.GetBalance_S(ctx, msg.Caller)

	// String corrections
	baseTokenData.Data.Symbol = bot.EscapeMarkdown(baseTokenData.Data.Symbol)
	baseTokenMeta.Description = bot.EscapeMarkdown(baseTokenMeta.Description)
	tokenBSymbol = bot.EscapeMarkdown(tokenBSymbol)

	// Manual escaping
	creatorBalanceStr := strings.Replace(strconv.FormatFloat(balance, 'f', 3, 64), ".", "\\.", 1) + " SOL"

	// Create the string of the liquidity
	liquidityStr := strconv.FormatFloat(msg.BaseMintLiquidity, 'f', 0, 64) + " " + baseTokenData.Data.Symbol + " / " + strconv.FormatFloat(msg.QuoteMintLiquidity, 'f', 1, 64) + " " + tokenBSymbol

	// Authority strings
	mintAuthStr := "🔴 *Enabled* 🔴"
	freezeAuthStr := "🔴 *Enabled* 🔴"
	if baseTokenData.MintAuthority == nil {
		mintAuthStr = "🟢 *Disabled* 🟢"
	}
	if baseTokenData.FreezeAuthority == nil {
		freezeAuthStr = "🟢 *Disabled* 🟢"
	}

	// Colour and emoji
	var costs float64 = 0
	var costsStr = "N/A ⚪"

	// Openbook info if available
	openbookInfo := openbook.GetOpenbookInfo(msg.BaseMint.String())
	if openbookInfo != nil {
		costs = openbookInfo.Costs
		if costs < 0.5 {
			costsStr = strings.Replace(strconv.FormatFloat(costs, 'f', 3, 64), ".", "\\.", 1) + " 🔴"
		} else if costs < 2 {
			costsStr = strings.Replace(strconv.FormatFloat(costs, 'f', 3, 64), ".", "\\.", 1) + " 🟠"
		} else {
			costsStr = strings.Replace(strconv.FormatFloat(costs, 'f', 3, 64), ".", "\\.", 1) + " 🟢"
		}
	}

	titleStr := bot.EscapeMarkdown("Pair: "+baseTokenData.Data.Symbol+" / "+tokenBSymbol+"\nCosts: "+costsStr+"\nLiquidity: "+liquidityStr) + "\nToken Mint Auth: " + mintAuthStr + "\nToken Freeze Auth: " + freezeAuthStr

	// Get top holder string
	topHoldersStr := getHolderString(ctx, msg.BaseMint, msg.PoolCoinTokenAccount, parsedSupply, msg.BaseMintLiquidity)

	var socialsStr string = ""
	if baseTokenMeta.Telegram != "" {
		socialsStr += "\nTelegram: " + utils.SocialtS(baseTokenMeta.Telegram)
	} else if baseTokenMeta.Twitter != "" {
		socialsStr += "\nTwitter: " + utils.SocialtS(baseTokenMeta.Twitter)
	} else if baseTokenMeta.Website != "" {
		socialsStr += "\nWebsite: " + utils.SocialtS(baseTokenMeta.Website)
	}
	if socialsStr != "" {
		socialsStr = bot.EscapeMarkdown(socialsStr)
		socialsStr = "\n\n*Socials*" + socialsStr
	}

	linkPreviewDisabled := false
	_, err := telegram.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatId,
		Text:   fmt.Sprintf("*\\[RAYDIUM POOL\\]*\n%s\n\n*Pair Address*\n`%s`\n*Token Address*\n`%s`\n*Creator Address* \\(%s\\)\n`%s`\n\n*Token Description*\n%s%s\n\n*Holders*\n%s", titleStr, msg.AmmID.String(), msg.BaseMint.String(), creatorBalanceStr, msg.Caller.String(), baseTokenMeta.Description, socialsStr, topHoldersStr),
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &linkPreviewDisabled,
		},
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Token - Solscan", URL: "https://solscan.io/account/" + msg.BaseMint.String()},
					{Text: "Market - Solscan", URL: "https://solscan.io/account/" + msg.AmmID.String()},
					{Text: "Tx - Solscan", URL: "https://solscan.io/tx/" + msg.TxID.String()},
				},
				{
					{Text: "RugCheck", URL: "https://rugcheck.xyz/tokens/" + msg.BaseMint.String()},
					{Text: "BirdEye", URL: "https://birdeye.so/token/" + msg.BaseMint.String()},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Error sending telegram message: %v\n", err)
	}

}
