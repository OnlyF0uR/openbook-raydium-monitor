package telegram_hook

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func tg_openbook_hook(msg *openbook.OpenbookInfo, ctx context.Context) {
	baseTokenData, baseTokenMeta := utils.TokenHelper(ctx, msg.BaseMint)
	if baseTokenData == nil || baseTokenMeta == nil {
		return
	}

	tokenBSymbol := utils.TokenToSymbol(msg.QuoteMint)

	balance := utils.GetBalance_S(ctx, msg.Caller)

	var titleEmoji = "🟢"
	if msg.Costs < 2 {
		titleEmoji = "🟠"
	}

	if msg.Costs < 0.5 {
		titleEmoji = "🔴"
	}

	// String corrections
	baseTokenData.Data.Symbol = bot.EscapeMarkdown(baseTokenData.Data.Symbol)
	baseTokenMeta.Description = bot.EscapeMarkdown(baseTokenMeta.Description)
	tokenBSymbol = bot.EscapeMarkdown(tokenBSymbol)

	costsStr := strings.Replace(strconv.FormatFloat(msg.Costs, 'f', 3, 64), ".", "\\.", 1)
	titleStr := "Pair: " + baseTokenData.Data.Symbol + " / " + tokenBSymbol + "\nCosts: " + costsStr + " SOL " + titleEmoji

	// Manual escaping
	creatorBalanceStr := strings.Replace(strconv.FormatFloat(balance, 'f', 3, 64), ".", "\\.", 1) + " SOL"

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
		Text:   fmt.Sprintf("*\\[OPENBOOK MARKET\\]*\n%s\n\n*Token Address*\n`%s`\n*Market Id*\n`%s`\n*Creator Address* \\(%s\\)\n`%s`\n\n*Token Description*\n%s%s", titleStr, msg.BaseMint.String(), msg.Market.String(), creatorBalanceStr, msg.Caller.String(), baseTokenMeta.Description, socialsStr),
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &linkPreviewDisabled,
		},
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Token - Solscan", URL: "https://solscan.io/account/" + msg.BaseMint.String()},
					{Text: "Market - Solscan", URL: "https://solscan.io/account/" + msg.Market.String()},
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
