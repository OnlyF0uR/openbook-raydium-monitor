package utils

import (
	"strconv"
	"strings"

	"github.com/gagliardetto/solana-go"
)

// StI64 converts a string to an int64
// and returns -1 if the conversion fails
func StI64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}
	return i
}

// I64tS converts an int64 to a string
func I64tS(i int64) string {
	return strconv.FormatInt(i, 10)
}

// SocialstS converts token meta socials to a formatted form
// and returns the string representation
func SocialstS(twitter string, telegram string, website string) string {
	socials := []string{}
	if twitter != "" {
		if !strings.HasPrefix(twitter, "https://") {
			twitter = "https://" + twitter
		}
		socials = append(socials, "[Twitter]("+twitter+")")
	}
	if telegram != "" {
		if !strings.HasPrefix(telegram, "https://") {
			telegram = "https://" + telegram
		}
		socials = append(socials, "[Telegram]("+telegram+")")
	}
	if website != "" {
		if !strings.HasPrefix(website, "https://") {
			website = "https://" + website
		}
		socials = append(socials, "[Website]("+website+")")
	}

	socialStr := strings.Join(socials, " | ")

	if socialStr == "" {
		socialStr = "None"
	}

	return socialStr
}

func SocialtS(social string) string {
	if social != "" {
		if !strings.HasPrefix(social, "https://") {
			social = "https://" + social
		}
	}

	return social
}

func TokenToSymbol(token solana.PublicKey) string {
	if token == solana.WrappedSol {
		return "SOL"
	} else if token == USDC_MINT_PUBKEY {
		return "USDC"
	} else {
		return "N/A"
	}
}
