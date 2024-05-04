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

func Test_GetRelatedTokens(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	tokens, err := GetRelatedTokens(ctx, solana.MustPublicKeyFromBase58("BoSAn75k8wvHoHL9Akkpk8oX4guvPLNfhLcKZ6zLqc8s"))
	if err != nil {
		t.Error(err)
	}

	t.Logf("tokens: %#+v", *tokens)
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

func Test_GetTokenAmount(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	amount := GetTokenAmount(ctx, solana.MustPublicKeyFromBase58("ANNiExyBjQ2iAViUbSFTvwWBvvVqQsE77QdAb4qxfTQj"), solana.WrappedSol)
	if amount == 0 {
		t.Error("amount is 0")
	}
	t.Logf("amount: %#+v", amount)
}
