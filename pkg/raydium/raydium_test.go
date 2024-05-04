package raydium

import (
	"context"
	"testing"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/gagliardetto/solana-go"
)

func Test_parseTransaction(t *testing.T) {
	ctx := context.Background()

	rpcs.Initialise([]string{})

	info := parseTransaction(ctx, solana.MustSignatureFromBase58("4iknGwBn1pxVgo5AMoRrgT4X4nnXoCcdrYYgkBypQkFqDtcthjbSnH1ijf8wwns95cCCzn8uY2VcE6sgWy8qbQf6"))
	t.Logf("info: %#+v", info)
}
