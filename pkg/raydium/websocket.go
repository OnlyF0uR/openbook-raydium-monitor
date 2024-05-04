package raydium

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func Start(ctx context.Context, wsUrl string, ch chan<- solana.Signature) error {
	client, err := ws.Connect(ctx, wsUrl)
	if err != nil {
		return err
	}

	fmt.Printf("Starting Raydium monitor\n")

	raydium := solana.MustPublicKeyFromBase58(utils.RAYDIUM_PROGRAM_ID)

	sub, err := client.LogsSubscribeMentions(
		raydium,
		rpc.CommitmentConfirmed,
	)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	var lastSignature string
	for {
		got, err := sub.Recv()
		if err != nil {
			return err
		}

		if got.Value.Signature.String() == lastSignature {
			continue
		}

		lastSignature = got.Value.Signature.String()

		if logFilter(got.Value.Logs) {
			ch <- got.Value.Signature
		}
	}
}

func logFilter(logs []string) bool {
	for _, log := range logs {
		if strings.Contains(log, utils.RAYDIUM_IDENTIFIER) {
			return true
		}
	}
	return false
}

func GetMetadata(logs []string) *json.RawMessage {
	for i := range logs {
		curLog := logs[i]

		// Parse IDO info from log
		_, after, found := strings.Cut(curLog, " InitializeInstruction2 ")
		if !found {
			continue // Search further, not IDO log.
		}

		// Add quotes to keys.
		splitted := strings.Split(after, " ")
		for i, s := range splitted {
			if strings.Contains(s, ":") {
				splitted[i] = "\"" + s[:len(s)-1] + "\":"
			}
		}

		metadata := json.RawMessage(strings.Join(splitted, " "))
		if !json.Valid(metadata) {
			continue // Search further, invalid JSON.
		}

		return &metadata
	}

	return nil
}
