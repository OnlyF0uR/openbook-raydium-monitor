package hooks

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/OnlyF0uR/solana-monitor/internal/load"
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

		channelID := os.Getenv("DISCORD_OPENBOOK_CHANNEL")
		if channelID == "" {
			fmt.Println("DISCORD_OPENBOOK_CHANNEL not set")
			return
		}

		// Required information about the involved tokens
		baseTokenData, quoteTokenData, baseTokenMeta := tokenHelper(ctx, msg.BaseMint, msg.QuoteMint)
		if baseTokenData == nil || quoteTokenData == nil || baseTokenMeta == nil {
			return
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
			titleEmoji = "🟠"
		}

		if msg.Costs < 0.5 {
			embedColour = utils.EMBED_COLOUR_RED
			titleEmoji = "🔴"
		}

		relatedTokensStr := getRelatedTokenString(ctx, msg.Caller, msg.BaseMint.String())

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Openbook hook timing (before discord: %v)\n", msg.TxID, time.Since(startTime))
		}

		fbSigner, fbAmount, err := utils.GetFundedBy_S(ctx, msg.Caller, msg.TxID)
		if err != nil {
			color.New(color.FgRed).Printf("Error getting funded by: %v\n", err)
			return
		}

		var fundedByAddress string
		if fbSigner == nil || fbAmount == 0 {
			color.New(color.FgYellow).Printf("No funded by found\n")
			fundedByAddress = ""
		} else {
			fundedByAddress = "\nFunded by: [" + fbSigner.Short(3) + "](https://solscan.io/account/" + fbSigner.String() + ") **(" + strconv.FormatFloat(fbAmount, 'f', 3, 64) + " SOL)**"
			fbName := load.FindFundedByFilter(fbSigner.String(), fbAmount)
			if fbName != "" {
				fundedByAddress = "\nFunded by: [" + fbName + "](https://solscan.io/account/" + fbSigner.String() + ") **(" + strconv.FormatFloat(fbAmount, 'f', 3, 64) + " SOL)**"
				warnings += "\n🚨 " + fbName + " detected 🚨"
			}
		}

		_, err = discord.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
			Title:       baseTokenData.Data.Symbol + "/" + quoteTokenData.Data.Symbol + " - " + strconv.FormatFloat(msg.Costs, 'f', 3, 64) + " SOL " + titleEmoji,
			Color:       embedColour,
			Description: warnings,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Addresses",
					Value:  "**Token**\n``" + msg.BaseMint.String() + "``\n**Market**\n``openbook:" + msg.Market.String() + "``" + fundedByAddress,
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
		})

		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
		}

		if os.Getenv("DEBUG") == "1" {
			spew.Dump(msg)
			color.New(color.FgBlue).Printf("[%s] Openbook hook timing (finished: %v)\n", msg.TxID, time.Since(startTime))
		}
	}

	fmt.Printf("Openbook hook out...\n")
}
