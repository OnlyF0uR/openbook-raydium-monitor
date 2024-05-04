package rpcs

import (
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/time/rate"
)

var rpcPool []*rpc.Client // Only reads, so thread safe?

func Initialise(rpcStrings []string) {
	for _, rpcString := range rpcStrings {
		client := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
			rpcString,
			rate.Every(time.Second), // time frame
			5,                       // limit of requests per time frame
		))
		rpcPool = append(rpcPool, client)
	}
	rpcPool = append(rpcPool, rpc.New(rpc.MainNetBeta_RPC))

	fmt.Printf("RPC pool initialised (total: %d)\n", len(rpcPool))
}

func BorrowClient(prevIndex int) (*rpc.Client, int) {
	if len(rpcPool) == 0 {
		panic("rpc pool not initialised")
	}

	index := (prevIndex + 1) % len(rpcPool)

	return rpcPool[index], index
}
