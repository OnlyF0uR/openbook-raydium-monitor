package raydium

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type RaydiumMetadata struct {
	Nonce          uint64 `json:"nonce"`
	OpenTime       uint64 `json:"open_time"`
	InitPcAmount   uint64 `json:"init_pc_amount"`
	InitCoinAmount uint64 `json:"init_coin_amount"`
}

type RaydiumInfo struct {
	// Initialize Market Instruction Data
	ProgramID            solana.PublicKey // always raydium
	AmmID                solana.PublicKey // Amm ID (Pair Address)
	AmmOpenOrders        solana.PublicKey // Amm Open Orders (PoolQuoteTokenAccount)
	LPTokenAddress       solana.PublicKey // LPToken Address (PoolTokenMint)
	BaseMint             solana.PublicKey // base mint address (Token Address)
	QuoteMint            solana.PublicKey // quote mint address (Currency Address)
	PoolCoinTokenAccount solana.PublicKey // Amm Token Account (PoolCoinTokenAccount)
	PoolPcTokenAccount   solana.PublicKey // Amm WSOL Token Account (PoolPcTokenAccount)
	AmmTargetOrders      solana.PublicKey // Amm Target Orders
	AmmLiquidityCreator  solana.PublicKey // Amm Liquidity Creator (aka account of LP creator that will receive LP tokens)

	BaseMintLiquidity  float64
	QuoteMintLiquidity float64

	// Initialize Market Instruction Metadata
	Caller    solana.PublicKey // Caller wallet address
	TxID      solana.Signature // Transaction ID
	Slot      uint64           // Chain Slot
	TxTime    time.Time        // Timestamp of transaction in blockchain
	Timestamp time.Time        // Timestamp of transaction discovery
	Swapped   bool             // Whether the pair was created in reverse order.

	Metadata RaydiumMetadata
}

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
