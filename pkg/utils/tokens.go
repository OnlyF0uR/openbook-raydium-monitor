package utils

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/fatih/color"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/yosefl20/solana-go-sdk/program/metaplex/token_metadata"
)

type Data struct {
	Name   string
	Symbol string
	Uri    string
}

type TokenData struct {
	// Optional authority used to mint new tokens. The mint authority may only be provided during
	// mint creation. If no mint authority is present then the mint has a fixed supply and no
	// further tokens may be minted.
	MintAuthority *solana.PublicKey

	// Total supply of tokens.
	Supply uint64

	// Number of base 10 digits to the right of the decimal place.
	Decimals uint8

	// Is `true` if this structure has been initialized
	IsInitialized bool

	// Optional authority to freeze token accounts.
	FreezeAuthority *solana.PublicKey

	Key                 uint8 // borsh.Enum
	UpdateAuthority     *solana.PublicKey
	Mint                *solana.PublicKey
	Data                Data
	PrimarySaleHappened bool
	IsMutable           bool
	EditionNonce        *uint8
	// TokenStandard       uint8 // borsh.Enum
	// Collection          *Collection
	// Uses                *Uses
	// CollectionDetails   *CollectionDetails
}

func getAccountData_S(ctx context.Context, tokenKey solana.PublicKey) *token.Mint {
	var mint *token.Mint
	lastRpcIndex := -1

	// Loop until we get a valid transaction
	iter := 0
	for {
		// If we fail to get the transaction after 6 attempts, break
		if iter == 6 {
			color.New(color.FgRed).Printf("getAccountData -> Failed to get account data after 6 attempts\n")
			return nil
		}
		if iter == 3 {
			time.Sleep(2 * time.Second)
		}

		client, lastClientIndex := rpcs.BorrowClient(lastRpcIndex)
		lastRpcIndex = lastClientIndex

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		err := client.GetAccountDataInto(wrapped_ctx, tokenKey, &mint)
		wrapped_cancel()

		if err == nil {
			// We found the account data
			break
		} else {
			if err.Error() == "not enough data: 15 bytes missing" {
				if os.Getenv("DEBUG") == "1" {
					color.New(color.FgRed).Printf("getAccountData -> Invalid mint account (missing 15 bytes): %s\n", tokenKey.String())
				}
				return nil
			} else {
				color.New(color.FgYellow).Printf("getAccountData -> Failed to get account data, retrying (%d): %v\n", lastRpcIndex+1, err)
			}
		}

		iter++
	}

	return mint
}

func getAccountInfo_S(ctx context.Context, metadataAccount solana.PublicKey) *rpc.GetAccountInfoResult {
	var accountInfo *rpc.GetAccountInfoResult
	lastRpcIndex := -1

	// Loop until we get a valid transaction
	iter := 0
	for {
		// If we fail to get the transaction after 6 attempts, break
		if iter == 6 {
			color.New(color.FgRed).Printf("getAccountInfo -> Failed to get account info after 5 attempts\n")
			return nil
		}

		client, lastClientIndex := rpcs.BorrowClient(lastRpcIndex)
		lastRpcIndex = lastClientIndex

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		tmp_accountInfo, err := client.GetAccountInfo(wrapped_ctx, metadataAccount)
		wrapped_cancel()

		if err == nil {
			// We found the account info
			accountInfo = tmp_accountInfo
			break
		} else {
			color.New(color.FgYellow).Printf("getAccountInfo -> Failed to get account info, retrying (%d): %v\n", lastRpcIndex+1, err)
		}

		iter++
	}

	return accountInfo
}

func GetTokendata(ctx context.Context, tokenKey solana.PublicKey, mayFail bool) (*TokenData, error) {
	var mint *token.Mint
	var metadataAccount solana.PublicKey
	var accountInfo *rpc.GetAccountInfoResult

	if mayFail {
		// No retries, just return error if failed
		client, _ := rpcs.BorrowClient(-1)

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		err := client.GetAccountDataInto(wrapped_ctx, tokenKey, &mint)
		wrapped_cancel()

		if err != nil {
			return nil, err
		}

		metadataAccountTmp, _, err := solana.FindTokenMetadataAddress(tokenKey)
		if err != nil {
			return nil, err
		}

		metadataAccount = metadataAccountTmp

		accountInfo, err = client.GetAccountInfo(ctx, metadataAccount)
		if err != nil {
			return nil, err
		}
	} else {
		// Includes 5 retries
		mint = getAccountData_S(ctx, tokenKey)
		if mint == nil {
			return nil, errors.New("failed to get mint account data")
		}

		metadataAccountTmp, _, err := solana.FindTokenMetadataAddress(tokenKey)
		if err != nil {
			return nil, err
		}
		metadataAccount = metadataAccountTmp

		accountInfo = getAccountInfo_S(ctx, metadataAccount)
		if accountInfo == nil {
			return nil, errors.New("failed to get metadata account info")
		}
	}

	data, err := token_metadata.MetadataDeserialize(accountInfo.Value.Data.GetBinary())
	if err != nil {
		return nil, err
	}

	// mintAddress := solana.MustPublicKeyFromBase58(data.Mint.String())
	updateAuthority := solana.MustPublicKeyFromBase58(data.UpdateAuthority.String())

	tokenData := TokenData{
		MintAuthority:   mint.MintAuthority,
		Supply:          mint.Supply,
		Decimals:        mint.Decimals,
		IsInitialized:   mint.IsInitialized,
		FreezeAuthority: mint.FreezeAuthority,
		Key:             uint8(data.Key),
		UpdateAuthority: &updateAuthority,
		Mint:            &tokenKey,
		Data: Data{
			Name:   data.Data.Name,
			Symbol: data.Data.Symbol,
			Uri:    data.Data.Uri,
		},
		PrimarySaleHappened: data.PrimarySaleHappened,
		IsMutable:           data.IsMutable,
		EditionNonce:        data.EditionNonce,
		// TokenStandard:       uint8(*data.TokenStandard), nil check required
	}

	return &tokenData, nil
}

type TokenMetaExtensions struct {
	Website  string `json:"website"`
	Twitter  string `json:"twitter"`
	Telegram string `json:"telegram"`
}

type TokenMeta struct {
	Name        string              `json:"name"`
	Symbol      string              `json:"symbol"`
	Description string              `json:"description"`
	Image       string              `json:"image"`
	ShowName    bool                `json:"show_name"`
	CreatedOn   string              `json:"created_on"`
	Twitter     string              `json:"twitter"`
	Telegram    string              `json:"telegram"`
	Website     string              `json:"website"`
	Extensions  TokenMetaExtensions `json:"extensions"`
}

func FetchTokenMeta(uri string) (*TokenMeta, error) {
	if strings.Contains(uri, "nftstorage.link") {
		splitted := strings.Split(uri, "/")
		splitted = strings.Split(splitted[len(splitted)-1], ".")

		uri = IPFS_GATEWAY + splitted[0]
	}

	var meta TokenMeta

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &meta)
	if err != nil {
		return nil, err
	}

	if meta.Extensions.Telegram != "" {
		meta.Telegram = meta.Extensions.Telegram
	}

	if meta.Extensions.Twitter != "" {
		meta.Twitter = meta.Extensions.Twitter
	}

	if meta.Extensions.Website != "" {
		meta.Website = meta.Extensions.Website
	}

	return &meta, nil
}

type TokenAccountMeta struct {
	Parsed struct {
		Info struct {
			IsNative    bool   `json:"isNative"`
			Mint        string `json:"mint"`
			Owner       string `json:"owner"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount         string  `json:"amount"`
				Decimals       int     `json:"decimals"`
				UIAmount       float64 `json:"uiAmount"`
				UIAmountString string  `json:"uiAmountString"`
			} `json:"tokenAmount"`
		} `json:"info"`
		Type string `json:"type"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   int    `json:"space"`
}

func GetRelatedTokens(ctx context.Context, account solana.PublicKey) (*[]TokenAccountMeta, error) {
	client, _ := rpcs.BorrowClient(-1)

	wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
	tokenAccounts, err := client.GetTokenAccountsByOwner(wrapped_ctx, account, &rpc.GetTokenAccountsConfig{
		ProgramId: &solana.TokenProgramID,
	}, &rpc.GetTokenAccountsOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   solana.EncodingJSONParsed,
	})
	wrapped_cancel()

	if err != nil {
		return nil, err
	}

	var relatedTokens []TokenAccountMeta
	for _, tokenAccount := range tokenAccounts.Value {
		if len(relatedTokens) > 6 {
			break
		}

		bytes, err := tokenAccount.Account.Data.MarshalJSON()
		if err != nil {
			continue
		}

		// Unmarshall
		var tokenAccountMeta TokenAccountMeta
		err = json.Unmarshal(bytes, &tokenAccountMeta)
		if err != nil {
			continue
		}

		relatedTokens = append(relatedTokens, tokenAccountMeta)
	}

	return &relatedTokens, nil
}

type TopHolder struct {
	PublicKey solana.PublicKey
	Amount    float64
}

func GetTopHolders_S(ctx context.Context, mint solana.PublicKey) *[]TopHolder {
	var rpcAccounts *rpc.GetTokenLargestAccountsResult
	lastRpcIndex := -1

	// Loop until we get a valid transaction
	iter := 0
	for {
		// If we fail to get the transaction after 6 attempts, break
		if iter == 6 {
			color.New(color.FgRed).Printf("getAccountInfo -> Failed to get account info after 5 attempts\n")
			return nil
		}

		client, lastClientIndex := rpcs.BorrowClient(lastRpcIndex)
		lastRpcIndex = lastClientIndex

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		tmp_accounts, err := client.GetTokenLargestAccounts(wrapped_ctx, mint, rpc.CommitmentConfirmed)
		wrapped_cancel()

		if err == nil {
			// We found the account info
			rpcAccounts = tmp_accounts
			break
		} else {
			color.New(color.FgYellow).Printf("getAccountInfo -> Failed to get account info, retrying (%d): %v\n", lastRpcIndex+1, err)
		}

		iter++
	}

	var topHolders []TopHolder
	for _, account := range rpcAccounts.Value {
		if account.UiTokenAmount.UiAmount == nil {
			continue
		}

		topHolders = append(topHolders, TopHolder{
			PublicKey: account.Address,
			Amount:    *account.UiTokenAmount.UiAmount, // .UiAmount is deprecated
		})
	}

	// Sort top holders by amount
	for i := 0; i < len(topHolders); i++ {
		for j := i + 1; j < len(topHolders); j++ {
			if topHolders[i].Amount < topHolders[j].Amount {
				topHolders[i], topHolders[j] = topHolders[j], topHolders[i]
			}
		}
	}

	return &topHolders
}

func GetTokenAmount(ctx context.Context, account solana.PublicKey, mint solana.PublicKey) float64 {
	client, _ := rpcs.BorrowClient(-1)

	// First we must get the token accounts
	wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
	result, err := client.GetTokenAccountsByOwner(wrapped_ctx, account, &rpc.GetTokenAccountsConfig{
		Mint: &mint,
	}, &rpc.GetTokenAccountsOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   solana.EncodingJSONParsed,
	})
	wrapped_cancel()

	if err != nil {
		return 0
	}

	if len(result.Value) == 0 {
		return 0
	}

	// Get the amount
	bytes, err := result.Value[0].Account.Data.MarshalJSON()
	if err != nil {
		return 0
	}

	var tokenAccountMeta TokenAccountMeta
	err = json.Unmarshal(bytes, &tokenAccountMeta)
	if err != nil {
		return 0
	}

	return tokenAccountMeta.Parsed.Info.TokenAmount.UIAmount
}
