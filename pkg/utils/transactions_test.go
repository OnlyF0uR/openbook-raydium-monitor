package utils

import (
	"context"
	"testing"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/gagliardetto/solana-go"
)

func TestGetFundedBy_S(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	signer, amount, err := GetFundedBy_S(ctx, solana.MustPublicKeyFromBase58("4n278nSSNX48Kui2zgLBTyWgvm2mmuZLFJbDRZMw9CWP"), solana.MustSignatureFromBase58("2XLW4xW7GkWBoxZbb8cvZG9UHoSiKhC5XWtqV8eBVJqgYBNgmKP8o5Ci3FicyjNiimwMHwwfRemwkBNzirEMnRx5"))
	if err != nil {
		t.Error(err)
	}

	if signer == nil {
		t.Error("signer is nil")
	}

	if amount == 0 {
		t.Error("amount is 0")
	}

	t.Logf("signer: %#+v\namount: %#+v", signer, amount)
}
