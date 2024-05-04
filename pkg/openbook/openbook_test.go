package openbook

import (
	"context"
	"testing"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/gagliardetto/solana-go"
)

func Test_parseTransaction(t *testing.T) {
	ctx := context.Background()

	rpcs.Initialise([]string{})

	info := parseTransaction(ctx, solana.MustSignatureFromBase58("3od1BuAnH6KY2qQA73LoLCT4t2aL2MC14MFxv4uVb4daMgwZzjgKW2HgYkGHHj4DFCa52Zuu42M8QRAeg4gR4v9k"))
	if info == nil {
		t.Errorf("info is nil")
	} else {
		t.Logf("info: %+v", info)
	}
}
