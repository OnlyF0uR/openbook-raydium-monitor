package utils

import (
	"context"
	"testing"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/gagliardetto/solana-go"
)

func Test_GetTokendata(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	tokenData, err := GetTokendata(ctx, solana.MustPublicKeyFromBase58("5dJyaVfERNXJ5PWxFfCsL2KsuqUZ9wUhxCr21ifxGMVi"), false)
	if err != nil {
		t.Error(err)
	}

	t.Logf("token: %#+v", tokenData)
}

func Test_FetchTokenMeta(t *testing.T) {
	res, err := FetchTokenMeta("https://bafybeicw3txn5yu3oscu4o5xkvzoajkydzom5zityuk7rxncmduodlj5m4.ipfs.cf-ipfs.com/")
	if err != nil {
		t.Error(err)
	}

	t.Logf("res: %#+v", res)
}

func Test_GetTopHolders(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	holders := GetTopHolders_S(ctx, solana.MustPublicKeyFromBase58("EoptP6e22xWGNYJCTGNS2A1S29Z3CKNPJJ6ASGq8yft6"))
	if holders == nil {
		t.Error("holders is nil")
	}
	t.Logf("holders: %#+v", holders)
}

func Test_TokenHelper(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	baseTokenData, baseTokenMeta := TokenHelper(ctx, solana.MustPublicKeyFromBase58("EoptP6e22xWGNYJCTGNS2A1S29Z3CKNPJJ6ASGq8yft6"))

	if baseTokenData == nil {
		t.Error("baseTokenData is nil")
	}

	if baseTokenMeta == nil {
		t.Error("baseTokenMeta is nil")
	}

	t.Logf("baseTokenData: %#+v", baseTokenData)
	t.Logf("baseTokenMeta: %#+v", baseTokenMeta)
}
