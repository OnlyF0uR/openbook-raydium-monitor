package hooks

import (
	"context"

	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/raydium"
)

var OpenbookHooks []func(*openbook.OpenbookInfo, context.Context)
var RaydiumHooks []func(*raydium.RaydiumInfo, context.Context)

func RegisterOpenbookHook(cb func(*openbook.OpenbookInfo, context.Context)) {
	OpenbookHooks = append(OpenbookHooks, cb)
}

func RegisterRaydiumHook(cb func(*raydium.RaydiumInfo, context.Context)) {
	RaydiumHooks = append(RaydiumHooks, cb)
}

func RunOpenbookHooks(ch <-chan *openbook.OpenbookInfo) {
	ctx := context.Background()
	for msg := range ch {
		// Loop through openbook hooks
		for _, v := range OpenbookHooks {
			v(msg, ctx)
		}
	}
}

func RunRaydiumHooks(ch <-chan *raydium.RaydiumInfo) {
	ctx := context.Background()
	for msg := range ch {
		// Loop through raydium hooks
		for _, v := range RaydiumHooks {
			v(msg, ctx)
		}
	}
}
