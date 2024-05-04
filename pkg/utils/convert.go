package utils

import (
	"strconv"
	"strings"
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
// and returns the string representation
func I64tS(i int64) string {
	return strconv.FormatInt(i, 10)
}

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
