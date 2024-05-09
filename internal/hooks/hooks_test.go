package hooks

import (
	"context"
	"testing"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/gagliardetto/solana-go"
)

func Test_tokenHelper(t *testing.T) {
	ctx := context.Background()
	rpcs.Initialise([]string{})

	baseTokenData, baseTokenMeta := tokenHelper(ctx, solana.MustPublicKeyFromBase58("EoptP6e22xWGNYJCTGNS2A1S29Z3CKNPJJ6ASGq8yft6"))

	if baseTokenData == nil {
		t.Error("baseTokenData is nil")
	}

	if baseTokenMeta == nil {
		t.Error("baseTokenMeta is nil")
	}

	t.Logf("baseTokenData: %#+v", baseTokenData)
	t.Logf("baseTokenMeta: %#+v", baseTokenMeta)
}
