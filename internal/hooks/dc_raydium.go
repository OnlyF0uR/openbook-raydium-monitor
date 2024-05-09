package hooks

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/raydium"
	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/gagliardetto/solana-go"
)

func RaydiumDiscord(ch <-chan *raydium.RaydiumInfo) {
	ctx := context.Background()
	for msg := range ch {
		startTime := time.Now()

		baseTokenData, baseTokenMeta := tokenHelper(ctx, msg.BaseMint)
		if baseTokenData == nil || baseTokenMeta == nil {
			continue
		}

		tokenBSymbol := tokenToSymbol(msg.QuoteMint)

		// Parse the supply to a float64
		parsedSupply := float64(baseTokenData.Supply) / math.Pow10(int(baseTokenData.Decimals))

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Raydium hook timing (before caller balance: %v)\n", msg.TxID, time.Since(startTime))
		}

		// Get the balance of the caller
		balance := utils.GetBalance_S(ctx, msg.Caller)

		// Create the string of the liquidity
		liquidityStr := strconv.FormatFloat(msg.BaseMintLiquidity, 'f', 0, 64) + " " + baseTokenData.Data.Symbol + " / " + strconv.FormatFloat(msg.QuoteMintLiquidity, 'f', 1, 64) + " " + tokenBSymbol

		// Authority strings
		mintAuthStr := "🔴 **Enabled** 🔴"
		freezeAuthStr := "🔴 **Enabled** 🔴"
		if baseTokenData.MintAuthority == nil {
			mintAuthStr = "🟢 **Disabled** 🟢"
		}
		if baseTokenData.FreezeAuthority == nil {
			freezeAuthStr = "🟢 **Disabled** 🟢"
		}

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Raydium hook timing (before related tokens: %v)\n", msg.TxID, time.Since(startTime))
		}

		// Related tokens
		relatedTokensStr := getRelatedTokenString(ctx, msg.Caller, msg.BaseMint.String())

		// Colour and emoji
		var embedColour = utils.EMBED_COLOUR_PURPLE
		var titleEmoji = "🟢"
		var costs float64 = 0

		// Openbook info if available
		openbookInfo := openbook.GetOpenbookInfo(msg.BaseMint.String())
		if openbookInfo != nil {
			costs = openbookInfo.Costs
		}

		if costs < 2 {
			embedColour = utils.EMBED_COLOUR_BLUE
		}

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Raydium hook timing (before warnings: %v)\n", msg.TxID, time.Since(startTime))
		}

		warnings := getWarningsString(ctx, msg.Caller, baseTokenMeta.CreatedOn)
		if warnings != "" {
			embedColour = utils.EMBED_COLOUR_ORANGE
		}

		if costs < 0.5 {
			embedColour = utils.EMBED_COLOUR_RED
			titleEmoji = "🔴"
		}

		if baseTokenMeta.CreatedOn == "https://pump.fun" {
			embedColour = utils.EMBED_COLOUR_GREEN
		}

		// Fallback for when no openbook information is available
		costsStr := "N/A ⚪"
		if costs > 0 {
			costsStr = strconv.FormatFloat(costs, 'f', 3, 64) + " SOL " + titleEmoji
		}

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Raydium hook timing (before top holders: %v)\n", msg.TxID, time.Since(startTime))
		}

		// Get top holder string
		topHoldersStr := getHolderString(ctx, msg.BaseMint, msg.PoolCoinTokenAccount, parsedSupply, msg.BaseMintLiquidity)

		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[%s] Raydium hook timing (before discord: %v)\n", msg.TxID, time.Since(startTime))
		}

		embed := &discordgo.MessageEmbed{
			Title:       baseTokenData.Data.Symbol + "/" + tokenBSymbol + " - " + costsStr,
			Color:       embedColour,
			Description: warnings,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Token Address",
					Value:  "``" + msg.BaseMint.String() + "``",
					Inline: false,
				},
				{
					Name:   "Pool Info",
					Value:  "Created: <t:" + utils.I64tS(msg.TxTime.Truncate(time.Second).Unix()) + ":R>\nOpens: <t:" + strconv.Itoa(int(msg.Metadata.OpenTime)) + ":R>\nCreator: [" + msg.Caller.Short(3) + "](https://solscan.io/account/" + msg.Caller.String() + ") **(" + strconv.FormatFloat(balance, 'f', 3, 64) + " SOL)**\nLiquidity: **" + liquidityStr + "**\nRelated Tokens: " + relatedTokensStr,
					Inline: false,
				},
				{
					Name:   "Token Description",
					Value:  baseTokenMeta.Description,
					Inline: false,
				},
				{
					Name:   "Authorities",
					Value:  "Mint: " + mintAuthStr + "\nFreeze: " + freezeAuthStr,
					Inline: true,
				},
				{
					Name:   "Token Ownership",
					Value:  topHoldersStr,
					Inline: true,
				},
				{
					Name:   "Socials",
					Value:  utils.SocialstS(baseTokenMeta.Twitter, baseTokenMeta.Telegram, baseTokenMeta.Website),
					Inline: false,
				},
				{
					Name:  "Extra Links",
					Value: "[Solscan (Token)](https://solscan.io/account/" + msg.BaseMint.String() + ") | [Solscan (Tx)](https://solscan.io/tx/" + msg.TxID.String() + ") | [Solscan (Pool)](https://solscan.io/account/" + msg.AmmID.String() + ") | [BirdEye](https://birdeye.so/token/" + msg.BaseMint.String() + ") | [RugCheck](https://rugcheck.xyz/tokens/" + msg.BaseMint.String() + ") | [Photon](https://photon-sol.tinyastro.io/en/lp/" + msg.AmmID.String() + ")",
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: baseTokenMeta.Image,
			},
		}

		_, err := discord.ChannelMessageSendEmbed(raydiumChannelID, embed)
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
			color.New(color.FgBlue).Printf("[%s] Raydium hook timing (finished: %v)\n", msg.TxID, time.Since(startTime))
		}
	}

	fmt.Printf("Raydium hook out...\n")
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

			var address = holder.PublicKey.Short(3)
			if holder.PublicKey == poolCoinTokenAccount {
				topHolderRaydiumAmount = "**Raydium: " + strconv.FormatFloat(supplyPct, 'f', 2, 64) + "%**\n\n"
				address = address + " (LP)"
			}
			if i < 5 {
				// We add to string here too
				topHoldersStr += "[" + address + "](https://solscan.io/account/" + holder.PublicKey.String() + ") - " + strconv.FormatFloat(supplyPct, 'f', 2, 64) + "%\n"
			}
		}
	}
	if len(topHoldersStr) > 0 {
		topHoldersStr = topHoldersStr[:len(topHoldersStr)-1]
	}

	if topHolderRaydiumAmount == "" {
		topHolderRaydiumAmount = "**Raydium: " + strconv.FormatFloat((liquidity/supply)*100, 'f', 2, 64) + "%**\n\n"
	}

	return topHolderRaydiumAmount + topHoldersStr
}
