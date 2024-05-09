package hooks

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

func OpenbookDiscord(ch <-chan *openbook.OpenbookInfo) {
	ctx := context.Background()
	for msg := range ch {
		startTime := time.Now()

		// Required information about the involved tokens
		baseTokenData, baseTokenMeta := tokenHelper(ctx, msg.BaseMint)
		if baseTokenData == nil || baseTokenMeta == nil {
			continue
		}

		tokenBSymbol := tokenToSymbol(msg.QuoteMint)

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Openbook hook timing (before caller balance: %v)\n", msg.TxID, time.Since(startTime))
		}

		balance := utils.GetBalance_S(ctx, msg.Caller)

		// Make description smaller if larger than 600 characters
		if len(baseTokenMeta.Description) > 600 {
			baseTokenMeta.Description = baseTokenMeta.Description[:600] + "..."
		}

		var embedColour = utils.EMBED_COLOUR_PURPLE
		var titleEmoji = "🟢"
		if msg.Costs < 2 {
			embedColour = utils.EMBED_COLOUR_BLUE
		}

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Openbook hook timing (before warnings: %v)\n", msg.TxID, time.Since(startTime))
		}

		warnings := getWarningsString(ctx, msg.Caller, baseTokenMeta.CreatedOn)
		if warnings != "" {
			embedColour = utils.EMBED_COLOUR_ORANGE
		}

		if msg.Costs < 0.5 {
			embedColour = utils.EMBED_COLOUR_RED
			titleEmoji = "🔴"
		}

		if baseTokenMeta.CreatedOn == "https://pump.fun" {
			embedColour = utils.EMBED_COLOUR_GREEN
		}

		relatedTokensStr := getRelatedTokenString(ctx, msg.Caller, msg.BaseMint.String())

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Openbook hook timing (before discord: %v)\n", msg.TxID, time.Since(startTime))
		}

		embed := &discordgo.MessageEmbed{
			Title:       baseTokenData.Data.Symbol + "/" + tokenBSymbol + " - " + strconv.FormatFloat(msg.Costs, 'f', 3, 64) + " SOL " + titleEmoji,
			Color:       embedColour,
			Description: warnings,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Addresses",
					Value:  "**Token**\n``" + msg.BaseMint.String() + "``\n**Market**\n``openbook:" + msg.Market.String() + "``",
					Inline: false,
				},
				{
					Name:   "Token",
					Value:  "Creator: [" + msg.Caller.Short(3) + "](https://solscan.io/account/" + msg.Caller.String() + ") **(" + strconv.FormatFloat(balance, 'f', 3, 64) + " SOL)**\nRelated Tokens: " + relatedTokensStr,
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
			fmt.Printf("Message: %v\n", msg)
		}

		_, err = discord.ChannelMessageSendEmbed(jointChannelID, embed)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			fmt.Printf("Message: %v\n", msg)
		}

		if os.Getenv("DEBUG") == "1" {
			spew.Dump(msg)
			color.New(color.FgBlue).Printf("[%s] Openbook hook timing (finished: %v)\n", msg.TxID, time.Since(startTime))
		}
	}

	fmt.Printf("Openbook hook out...\n")
}
