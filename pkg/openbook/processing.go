package openbook

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/OnlyF0uR/solana-monitor/pkg/utils"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func ProcessMessages(rChn <-chan solana.Signature, sendChn chan<- *OpenbookInfo) {
	ctx := context.Background()

	for msg := range rChn {
		info := parseTransaction(ctx, msg)
		if info == nil {
			continue
		}

		SetOpenbookInfo(info.BaseMint.String(), info)
		sendChn <- info
	}

	fmt.Printf("Openbook processing out...\n")
}

func parseTransaction(ctx context.Context, signature solana.Signature) *OpenbookInfo {
	rpcTx, tx, err := utils.GetConfirmedTransaction_S(ctx, signature)
	if err != nil {
		fmt.Printf("Openbook -> parseTransaction error.\nhttps://solscan.io/tx/%s\n", signature.String())
		return nil
	}

	if tx == nil || rpcTx == nil {
		fmt.Printf("Openbook: failed to get transaction: %s\n", signature.String())
		return nil
	}

	if (len(tx.Message.Instructions)) < 6 {
		return nil
	}

	for _, instr := range tx.Message.Instructions {
		program, err := tx.Message.Program(instr.ProgramIDIndex)
		if err != nil {
			continue // Program account index out of range.
		}

		if program.String() != utils.OPENBOOK_PRGRAM_ID {
			continue // Not called by openbook.
		}

		if len(instr.Accounts) < 10 {
			continue // Not enough accounts for InitializeMarket instruction.
		}

		info := OpenbookInfo{
			ProgramID: program,
		}
		wasFound := destructInfo(instr, rpcTx, tx, &info)

		if wasFound {
			// Set the costs (this should be safe)
			info.Costs = (float64(rpcTx.Meta.PreBalances[0]) - float64(rpcTx.Meta.PostBalances[0])) / float64(solana.LAMPORTS_PER_SOL)
			return &info
		}
	}

	return nil
}

func destructInfo(instr solana.CompiledInstruction, rpcTx *rpc.GetTransactionResult, tx *solana.Transaction, info *OpenbookInfo) bool {
	const BaseMintIndex = 7
	const QuoteMinIndex = 8

	addressList := rpcTx.Meta.LoadedAddresses.Writable
	addressList = append(addressList, rpcTx.Meta.LoadedAddresses.ReadOnly...)

	if len(addressList) >= 9 {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("Using fallback address retrieval method for: %s\n", tx.Signatures[0].String())
		}

		if addressList[QuoteMinIndex] != solana.WrappedSol && addressList[BaseMintIndex] != solana.WrappedSol {
			// fmt.Printf("Openbook: found openbook market, but not with SOL currency\n")
			return false
		}

		// Set values
		info.Market = addressList[0]
		info.EventQueue = addressList[2]
		info.Bids = addressList[3]
		info.Asks = addressList[4]
		info.BaseVault = addressList[5]
		info.QuoteVault = addressList[6]
		info.BaseMint = addressList[7]
		info.QuoteMint = addressList[8]
	} else {
		safeIndex := func(idx uint16) solana.PublicKey {
			if idx >= uint16(len(tx.Message.AccountKeys)) {
				return solana.PublicKey{}
			}
			return tx.Message.AccountKeys[idx]
		}

		// if safeIndex(instr.Accounts[QuoteMinIndex]) != solana.WrappedSol && safeIndex(instr.Accounts[BaseMintIndex]) != solana.WrappedSol {
		// 	fmt.Printf("Openbook: found openbook market, but not with SOL currency\n")
		// 	return false
		// }

		// Set values
		info.Market = safeIndex(instr.Accounts[0])
		info.EventQueue = safeIndex(instr.Accounts[2])
		info.Bids = safeIndex(instr.Accounts[3])
		info.Asks = safeIndex(instr.Accounts[4])
		info.BaseVault = safeIndex(instr.Accounts[5])
		info.QuoteVault = safeIndex(instr.Accounts[6])
		info.BaseMint = safeIndex(instr.Accounts[7])
		info.QuoteMint = safeIndex(instr.Accounts[8])
	}

	if info.BaseMint == solana.SystemProgramID || info.QuoteMint == solana.SystemProgramID {
		return false
	}

	if info.BaseMint == solana.WrappedSol {
		info.BaseVault, info.QuoteVault = info.QuoteVault, info.BaseVault
		info.BaseMint, info.QuoteMint = info.QuoteMint, info.BaseMint
		info.Swapped = true
	} else if info.BaseMint.String() == utils.USDC_MINT {
		info.BaseMint, info.QuoteMint = info.QuoteMint, info.BaseMint
		info.BaseVault, info.QuoteVault = info.QuoteVault, info.BaseVault
		info.Swapped = true
	}

	info.Caller = tx.Message.AccountKeys[0] // Should be ok, but not sure.
	info.TxID = tx.Signatures[0]

	vaultSignerNonce := instr.Data[23:31]
	vaultsigner, err := solana.CreateProgramAddress(
		[][]byte{
			info.Market.Bytes()[:],
			vaultSignerNonce,
		}, solana.MustPublicKeyFromBase58(utils.OPENBOOK_PRGRAM_ID))

	if err != nil {
		fmt.Printf("Openbook: failed to create vault signer: %v\n", err)
		return false
	}

	info.VaultSigner = vaultsigner

	info.Slot = rpcTx.Slot
	info.TxTime = rpcTx.BlockTime.Time()
	info.Timestamp = time.Now()

	return true
}
