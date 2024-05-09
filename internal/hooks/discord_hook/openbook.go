package discord_hook

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

func dc_openbook_hook(msg *openbook.OpenbookInfo, ctx context.Context) {
	startTime := time.Now()

	// Required information about the involved tokens
	baseTokenData, baseTokenMeta := utils.TokenHelper(ctx, msg.BaseMint)
	if baseTokenData == nil || baseTokenMeta == nil {
		return
	}

	tokenBSymbol := utils.TokenToSymbol(msg.QuoteMint)

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[%s] Openbook hook timing (before caller balance: %v)\n", msg.TxID, time.Since(startTime))
	}

	balance := utils.GetBalance_S(ctx, msg.Caller)

	var embedColour = utils.EMBED_COLOUR_PURPLE
	var titleEmoji = "🟢"
	if msg.Costs < 2 {
		embedColour = utils.EMBED_COLOUR_BLUE
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[%s] Openbook hook timing (before warnings: %v)\n", msg.TxID, time.Since(startTime))
	}

	if msg.Costs < 0.5 {
		embedColour = utils.EMBED_COLOUR_RED
		titleEmoji = "🔴"
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[%s] Openbook hook timing (before discord: %v)\n", msg.TxID, time.Since(startTime))
	}

	embed := &discordgo.MessageEmbed{
		Title: baseTokenData.Data.Symbol + "/" + tokenBSymbol + " - " + strconv.FormatFloat(msg.Costs, 'f', 3, 64) + " SOL " + titleEmoji,
		Color: embedColour,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Addresses",
				Value:  "**Token**\n``" + msg.BaseMint.String() + "``\n**Market**\n``" + msg.Market.String() + "``",
				Inline: false,
			},
			{
				Name:   "Token",
				Value:  "Creator: [" + msg.Caller.Short(3) + "](https://solscan.io/account/" + msg.Caller.String() + ") **(" + strconv.FormatFloat(balance, 'f', 3, 64) + " SOL)**",
				Inline: true,
			},
			{
				Name:   "History",
				Value:  "Created: <t:" + utils.I64tS(msg.TxTime.Truncate(time.Second).Unix()) + ":R>",
				Inline: true,
			},
			{
				Name:   "Token Description",
				Value:  baseTokenMeta.Description,
				Inline: false,
			},
			{
				Name:   "Socials",
				Value:  utils.SocialstS(baseTokenMeta.Twitter, baseTokenMeta.Telegram, baseTokenMeta.Website),
				Inline: false,
			},
			{
				Name:  "Extra Links",
				Value: "[Solscan (Token)](https://solscan.io/account/" + msg.BaseMint.String() + ") | [Solscan (Tx)](https://solscan.io/tx/" + msg.TxID.String() + ") | [BirdEye](https://birdeye.so/token/" + msg.BaseMint.String() + ") | " + "[RugCheck](https://rugcheck.xyz/tokens/" + msg.BaseMint.String() + ")",
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: baseTokenMeta.Image,
		},
	}

	_, err := discord.ChannelMessageSendEmbed(openbookChannelID, embed)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}

	if os.Getenv("DEBUG") == "1" {
		spew.Dump(msg)
		color.New(color.FgBlue).Printf("[%s] Openbook hook timing (finished: %v)\n", msg.TxID, time.Since(startTime))
	}
}
