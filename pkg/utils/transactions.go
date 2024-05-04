package utils

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/davecgh/go-spew/spew"
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
	lastRpcIndex := -1

	// Loop until we get a valid transaction
	iter := 0
	for {
		// If we fail to get the transaction after 6 attempts, break
		if iter == 6 {
			color.New(color.FgRed).Printf("getRpcTx -> Failed to get transaction after 5 attempts\n")
			return nil, nil, errors.New("failed to get transaction after 5 attempts")
		}

		client, lastClientIndex := rpcs.BorrowClient(lastRpcIndex)
		lastRpcIndex = lastClientIndex

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		tmp_rpcTx, err := client.GetTransaction(wrapped_ctx, signature, &rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &Max_Transaction_Version,
			Commitment:                     rpc.CommitmentConfirmed,
			Encoding:                       solana.EncodingBase64,
		})
		wrapped_cancel()

		if err == nil {
			// We found the transaction
			rpcTx = tmp_rpcTx
			break
		} else {
			color.New(color.FgYellow).Printf("getRpcTx -> Failed to get transaction, retrying (%d): %v\n", lastRpcIndex+1, err)
		}

		iter++
	}

	// if rpcTx.Meta.Err != nil {
	// 	return nil, nil, fmt.Errorf("transaction failed: %v", rpcTx.Meta.Err)
	// }

	tx, err := rpcTx.Transaction.GetTransaction()
	if err != nil {
		color.New(color.FgRed).Printf("GetConfirmedTransaction -> Failed to get transaction: %v\n", err)
		return nil, nil, err
	}

	return rpcTx, tx, nil
}

func GetBalance_S(ctx context.Context, account solana.PublicKey) float64 {
	var balance float64 = 0
	lastRpcIndex := -1

	// Loop until we get a valid transaction
	iter := 0
	for {
		// If we fail to get the transaction after 6 attempts, break
		if iter == 6 {
			color.New(color.FgRed).Printf("getBalance -> Failed to get balance after 5 attempts\n")
			break
		}

		client, lastClientIndex := rpcs.BorrowClient(lastRpcIndex)
		lastRpcIndex = lastClientIndex

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 5*time.Second)
		result, err := client.GetBalance(wrapped_ctx, account, rpc.CommitmentConfirmed)
		wrapped_cancel()

		if err == nil {
			// We found the balance
			balance = float64(result.Value) / float64(solana.LAMPORTS_PER_SOL)
			break
		} else {
			color.New(color.FgYellow).Printf("getBalance -> Failed to get balance, retrying (%d): %v\n", lastRpcIndex+1, err)
		}

		iter++
	}

	return balance
}

func GetFundedBy_S(ctx context.Context, account solana.PublicKey, lastSignature solana.Signature) (*solana.PublicKey, float64, error) {
	lastRpcIndex := -1

	// Loop until we get a valid transaction
	iter := 0
	for {
		// If we fail to get the transaction after 6 attempts, break
		if iter == 6 {
			color.New(color.FgRed).Printf("getFundedBy -> Failed to get funded by after 5 attempts\n")
			return nil, 0, errors.New("failed to get funded by after 5 attempts")
		}

		client, lastClientIndex := rpcs.BorrowClient(lastRpcIndex)
		lastRpcIndex = lastClientIndex

		wrapped_ctx, wrapped_cancel := context.WithTimeout(ctx, 10*time.Second)
		result, err := client.GetSignaturesForAddressWithOpts(wrapped_ctx, account, &rpc.GetSignaturesForAddressOpts{
			// Before:     lastSignature,
			Commitment: rpc.CommitmentConfirmed,
		})
		wrapped_cancel()

		if err == nil {
			if len(result) == 0 {
				if os.Getenv("DEBUG") == "1" {
					spew.Dump(account)
					spew.Dump(result)
				}
				return nil, 0, errors.New("no transactions found")
			}

			// We found the balance
			// return result.Value, nil
			oldestTx := result[len(result)-1]

			rpcTx, tx, err := GetConfirmedTransaction_S(ctx, oldestTx.Signature)
			if err != nil {
				color.New(color.FgRed).Printf("getFundedBy -> Failed to get transaction: %v\n", err)
				return nil, 0, err
			}

			signer := tx.Message.AccountKeys[0]

			// Loop through prebalances
			for i, bal := range rpcTx.Meta.PreBalances {
				if bal != 0 {
					continue
				}

				newBal := rpcTx.Meta.PostBalances[i]
				if newBal == 0 {
					continue
				}

				return &signer, float64(newBal) / float64(solana.LAMPORTS_PER_SOL), nil
			}

			return nil, 0, nil
		} else {
			color.New(color.FgYellow).Printf("getFundedBy -> Failed to get funded by, retrying (%d): %v\n", lastRpcIndex+1, err)
		}

		iter++
	}
}
