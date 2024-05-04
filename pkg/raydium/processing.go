package raydium

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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

func ProcessMessages(rChn <-chan solana.Signature, sendChn chan<- *RaydiumInfo) {
	ctx := context.Background()

	for msg := range rChn {
		info := parseTransaction(ctx, msg)
		if info == nil {
			continue
		}

		sendChn <- info
	}

	fmt.Printf("Raydium processing out...\n")
}

func parseTransaction(ctx context.Context, signature solana.Signature) *RaydiumInfo {
	rpcTx, tx, err := utils.GetConfirmedTransaction_S(ctx, signature)
	if err != nil {
		fmt.Printf("Raydium -> parseTransaction: %v\nhttps://solscan.io/tx/%s\n", err, signature.String())
		return nil
	}

	if rpcTx.Meta.Err != nil {
		// fmt.Printf("Transaction failed: %v\nhttps://solscan.io/tx/%s\n", rpcTx.Meta.Err, signature)
		return nil
	}

	for _, instr := range tx.Message.Instructions {
		program, err := tx.Message.Program(instr.ProgramIDIndex)
		if err != nil {
			continue // Program account index out of range.
		}

		if program.String() != utils.RAYDIUM_PROGRAM_ID {
			continue // Not called by raydium.
		}

		// Now we know this is a raydium market.

		minfo := RaydiumInfo{
			ProgramID: program,
		}
		wasFound := destructInfo(instr, rpcTx, tx, &minfo)

		if wasFound {
			metadata := GetMetadata(rpcTx.Meta.LogMessages)
			bytes, err := metadata.MarshalJSON()
			if err != nil {
				return nil
			}

			var metadataStruct RaydiumMetadata
			err = json.Unmarshal(bytes, &metadataStruct)
			if err != nil {
				return nil
			}

			if metadataStruct.OpenTime == 0 {
				metadataStruct.OpenTime = uint64(minfo.Timestamp.Unix())
			}

			minfo.Metadata = metadataStruct

			return &minfo
		}
	}

	return nil
}

func destructInfo(instr solana.CompiledInstruction, rpcTx *rpc.GetTransactionResult, tx *solana.Transaction, info *RaydiumInfo) bool {
	safeIndex := func(idx uint16) solana.PublicKey {
		if idx >= uint16(len(tx.Message.AccountKeys)) {
			return solana.PublicKey{}
		}
		return tx.Message.AccountKeys[idx]
	}

	const BaseMintIndex = 8
	const QuoteMinIndex = 9

	// if safeIndex(instr.Accounts[QuoteMinIndex]) != solana.WrappedSol && safeIndex(instr.Accounts[BaseMintIndex]) != solana.WrappedSol {
	// 	color.New(color.FgYellow).Println("Raydium: found raydium market, but not with SOL currency")
	// 	return false
	// }

	info.AmmID = safeIndex(instr.Accounts[4])
	info.AmmOpenOrders = safeIndex(instr.Accounts[6])
	info.LPTokenAddress = safeIndex(instr.Accounts[7])
	info.BaseMint = safeIndex(instr.Accounts[BaseMintIndex])
	info.QuoteMint = safeIndex(instr.Accounts[QuoteMinIndex])
	info.PoolCoinTokenAccount = safeIndex(instr.Accounts[10])
	info.PoolPcTokenAccount = safeIndex(instr.Accounts[11])
	info.AmmTargetOrders = safeIndex(instr.Accounts[12])
	info.AmmLiquidityCreator = safeIndex(instr.Accounts[20])

	// Loop through posttokenbalances, find where owner is the raydium auth (5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1) and get the amount
	for _, postBalance := range rpcTx.Meta.PostTokenBalances {
		if postBalance.Owner.String() == utils.RAYDIUM_AUTHORITY_ID {
			if postBalance.Mint.Equals(info.BaseMint) {
				// .UiAmount is deprecated
				info.BaseMintLiquidity = *postBalance.UiTokenAmount.UiAmount
			} else if postBalance.Mint.Equals(info.QuoteMint) {
				// .UiAmount is deprecated
				info.QuoteMintLiquidity = *postBalance.UiTokenAmount.UiAmount
			}
		}
	}

	if info.BaseMint == solana.SystemProgramID || info.QuoteMint == solana.SystemProgramID {
		return false
	}

	if info.BaseMint == solana.WrappedSol {
		info.BaseMint, info.QuoteMint = info.QuoteMint, info.BaseMint
		info.BaseMintLiquidity, info.QuoteMintLiquidity = info.QuoteMintLiquidity, info.BaseMintLiquidity
		info.PoolCoinTokenAccount, info.PoolPcTokenAccount = info.PoolPcTokenAccount, info.PoolCoinTokenAccount
		info.Swapped = true
	} else if info.BaseMint.String() == utils.USDC_MINT {
		info.BaseMint, info.QuoteMint = info.QuoteMint, info.BaseMint
		info.BaseMintLiquidity, info.QuoteMintLiquidity = info.QuoteMintLiquidity, info.BaseMintLiquidity
		info.PoolCoinTokenAccount, info.PoolPcTokenAccount = info.PoolPcTokenAccount, info.PoolCoinTokenAccount
		info.Swapped = true
	}

	info.Caller = tx.Message.AccountKeys[0] // Should be ok, but not sure.
	info.TxID = tx.Signatures[0]

	info.Slot = rpcTx.Slot
	info.TxTime = rpcTx.BlockTime.Time()
	info.Timestamp = time.Now()

	return true
}
