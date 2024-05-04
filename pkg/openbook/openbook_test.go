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

	txSlice := []solana.Signature{
		solana.MustSignatureFromBase58("3od1BuAnH6KY2qQA73LoLCT4t2aL2MC14MFxv4uVb4daMgwZzjgKW2HgYkGHHj4DFCa52Zuu42M8QRAeg4gR4v9k"),
		solana.MustSignatureFromBase58("2JNHSYYMyAs6PQRcbxgQLMAvKUc25gYvMbNxKbxL944bUwTEpB3AihavSUCHvi4KnKb7Fux3NxRWzx4rtXUnb5Bu"),
		solana.MustSignatureFromBase58("4Fnur7g1ZgUxr5eBxRMpcBQC1ukm9EAeTbpwiKN3RrUUGhcgSg3PKSt9cW2i4FRyWYmTXJrB8Xg4XYYrn6Eqj9oA"),
		solana.MustSignatureFromBase58("2sjsMQQyrNzj3JMbwWPS7xrgrBnzXdRsoGJ5DDnwSu9kB8W1PNQUnhiUFrCwKsRkTDZ8V3FRcREvk7ja7X16WAZm"),
		solana.MustSignatureFromBase58("KHbVHghuMYcQsDWoLc4vP7GpFMa18UmTy7T9Kpd9amB8mceW4G3Jb1M5KeZCammYvEHs7mTp7d73JRHoAkckPgF"),
		solana.MustSignatureFromBase58("2xz4RjZaCyBQqXHX7upyLo8EYrsbZPj7pC7Er1fZMH3sEGoTpwMsvqQYBiKKvGB4A9KMCwYz7jNobXrPyoNoQBoa"),
		solana.MustSignatureFromBase58("3rQHwwFrMv2UEbJ31xEJs5pXX8FDsMV2PaQTPhC5k1Y1yVA4QUD5tHN5N7ZWb41b7e8LuhXBbDw5KMDxHm7XvANK"),
		solana.MustSignatureFromBase58("24De88c23XMJgbzzJFQe1raUhkPMdHZqRDfr3Ye11Y5iXWrp1qw7CYcEpWSbWFqJZckJ1JdAvbWYpFChZZZdsanh"),
		solana.MustSignatureFromBase58("4NNXs2g3zYtuYTh2wptKg5shUNEBA4C2jMqZQmGYkfTpuyrs6oVapEtUeLz5Viqsg1kFavvyKe3P9sV4o7anvbcS"),
		solana.MustSignatureFromBase58("2N8LXhvVXn6NxYp41VCuCz66ojp7pSeCoPvhkkYdi4TSqSKkHqVwQhqrRTmMmPeHWy5Ryz31Ewiej27TZQKy2Xmt"),
		solana.MustSignatureFromBase58("GsT7M1W8SCZQoPoNoAkg8Gc1vaSkLGCigNeP1kkeoPfSN1zVXLBSKw39S2dj5rwTdcnQ1Ne7YfUePJ4dZ51Qq1k"),
		solana.MustSignatureFromBase58("ujXJZPBNNhM8UDqVQwGJ3UUq14xXhrWHfCwbgMr1grMkPgwcJhskJdxNEQrER76zbmtGoSVr9Q8QgW21RSVAkdE"),
		solana.MustSignatureFromBase58("4gLo5uw5JF8SNvcxQw16TE8YfrA3dkbVcnvFx4PJiTXXDXGRD9yhWVKZMknN79vEwc1gACTcz6BDb3n99VRxSD5L"),
	}

	for _, tx := range txSlice {
		info := parseTransaction(ctx, tx)

		if info == nil {
			t.Errorf("info is nil: https://solscan.io/tx/%s", tx.String())
		} else {
			t.Logf("info: %+v", info)
		}
	}
}
