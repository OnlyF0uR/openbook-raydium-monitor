package openbook

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type OpenbookInfo struct {
	// Initialize Market Instruction Data
	ProgramID  solana.PublicKey // always OpenBook v3
	Market     solana.PublicKey // serum market address
	EventQueue solana.PublicKey // serum event queue address
	Bids       solana.PublicKey // serum bids address
	Asks       solana.PublicKey // serum asks address
	BaseMint   solana.PublicKey // base mint address (Token Address)
	QuoteMint  solana.PublicKey // quote mint address (Currency Address)
	BaseVault  solana.PublicKey // base vault address (Token Account)
	QuoteVault solana.PublicKey // quote vault address (Currency Account)

	// Initialize Market Instruction Metadata
	Caller    solana.PublicKey // Caller wallet address
	TxID      solana.Signature // Transaction ID
	Slot      uint64           // Chain Slot
	TxTime    time.Time        // Timestamp of transaction in blockchain
	Timestamp time.Time        // Timestamp of transaction discovery
	Swapped   bool             // Whether the pair was created in reverse order.

	// Initialize Market Instruction Extra
	VaultSigner solana.PublicKey // Vault signer; this value is provided by separate RPC call.

	Costs float64 // Costs of openbook creation
}

func Start(ctx context.Context, wsUrl string, ch chan<- solana.Signature) error {
	client, err := ws.Connect(ctx, wsUrl)
	if err != nil {
		return err
	}

	fmt.Printf("Starting Openbook monitor\n")

	openbook := solana.MustPublicKeyFromBase58(utils.OPENBOOK_PRGRAM_ID)

	sub, err := client.LogsSubscribeMentions(
		openbook,
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
	for i, log := range logs {
		if !strings.Contains(log, "Program 11111111111111111111111111111111 success") {
			continue // Search further.
		}
		if i+1 >= len(logs) {
			break // No more logs.
		}

		nextLog := logs[i+1]
		if !strings.Contains(nextLog, "Program srmqPvymJeFKQ4zGQed1GFppgkRHL9kaELCbyksJtPX invoke [1]") {
			continue // Search further.
		}

		// Found it
		return true
	}
	return false
}
