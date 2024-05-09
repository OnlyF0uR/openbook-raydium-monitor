package utils

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/fatih/color"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var Max_Transaction_Version uint64 = 0

type FetchTransactionResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  *rpc.GetTransactionResult
	Id      int `json:"id"`
}

// Get a confirmed transaction, will retry with mainnet beta rpc if failed, includes logging by default
func GetConfirmedTransaction_S(ctx context.Context, signature solana.Signature) (*rpc.GetTransactionResult, *solana.Transaction, error) {
	var rpcTx *rpc.GetTransactionResult

	for i := 0; i < 5; i++ {
		client := rpcs.BorrowClient()

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		tmp_rpcTx, err := client.GetTransaction(wrapped_ctx, signature, &rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &Max_Transaction_Version,
			Commitment:                     rpc.CommitmentConfirmed,
			Encoding:                       solana.EncodingBase64,
		})
		wrapped_cancel()

		if err != nil {
			if !strings.Contains(err.Error(), "context deadline exceeded") {
				if os.Getenv("DEBUG") == "1" {
					color.New(color.FgYellow).Printf("GetConfirmedTransaction_S -> Failed to get transaction, retrying (%d): %v\n", i+1, err)
				}
			}
			continue
		}

		rpcTx = tmp_rpcTx
		break
	}

	if rpcTx == nil {
		color.New(color.FgRed).Printf("GetConfirmedTransaction_S -> Failed to get transaction after 5 attempts\n")
		return nil, nil, errors.New("failed to get transaction after 5 attempts")
	}

	tx, err := rpcTx.Transaction.GetTransaction()
	if err != nil {
		color.New(color.FgRed).Printf("GetConfirmedTransaction_S -> Failed to get transaction: %v\n", err)
		return nil, nil, err
	}

	return rpcTx, tx, nil
}

func GetBalance_S(ctx context.Context, account solana.PublicKey) float64 {
	for i := 0; i < 5; i++ {
		client := rpcs.BorrowClient()
		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		result, err := client.GetBalance(wrapped_ctx, account, rpc.CommitmentConfirmed)
		wrapped_cancel()

		if err != nil {
			color.New(color.FgYellow).Printf("getBalance -> Failed to get balance, retrying (%d): %v\n", i+1, err)
			continue
		}

		return float64(result.Value) / float64(solana.LAMPORTS_PER_SOL)
	}

	// If we fail to get the balance after 5 attempts, return 0
	return 0
}
