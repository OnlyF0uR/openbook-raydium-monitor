package rpcs

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/time/rate"
)

var rpcPool []*rpc.Client // Only reads, so thread safe?
var rpcIndex int = 0

var mutex = &sync.Mutex{}

func Initialise(rpcStrings []string) {
	for _, rpcString := range rpcStrings {
		client := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
			rpcString,
			rate.Every(time.Second), // time frame
			4,                       // limit of requests per time frame
		))
		rpcPool = append(rpcPool, client)
	}

	if os.Getenv("INCLUDE_SOLANA_BETA_MAINNET_RPC") == "1" {
		rpcPool = append(rpcPool, rpc.New(rpc.MainNetBeta_RPC))
	}

	fmt.Printf("RPC pool(s) initialised (total: %d)\n", len(rpcPool))
}

func BorrowClient() *rpc.Client {
	if len(rpcPool) == 0 {
		panic("no RPC clients configured")
	}

	mutex.Lock()
	defer mutex.Unlock()

	if rpcIndex+1 == len(rpcPool) {
		rpcIndex = 0
	} else {
		rpcIndex++
	}

	return rpcPool[rpcIndex]
}
