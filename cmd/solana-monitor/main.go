package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/OnlyF0uR/solana-monitor/internal/hooks"
	"github.com/OnlyF0uR/solana-monitor/internal/load"
	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/raydium"
	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
)

func main() {
	// Print a welcome message including the version, build date, and developer
	fmt.Println("====================================")
	fmt.Println("Welcome to Solana Monitor")
	fmt.Println("Version: 1.0.1")
	fmt.Println("Build Date: 2024-05-04")
	fmt.Println("Developer: OnlyF0uR")
	fmt.Println("====================================")

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	load.LoadFundedByFilters()

	rpcUrl := os.Getenv("SOLANA_RPC_URL")
	rpcUrl2 := os.Getenv("SOLANA_RPC_URL_BACKUP")

	wsUrl := os.Getenv("SOLANA_WS_URL")

	// Channels for processing
	raydiumProcessingCh := make(chan solana.Signature)
	openbookProcessingCh := make(chan solana.Signature)
	// Channels for hooks
	raydiumHookCh := make(chan *raydium.RaydiumInfo)
	openbookHookCh := make(chan *openbook.OpenbookInfo)

	rpcs.Initialise([]string{
		rpcUrl,
		rpcUrl2,
	})

	var wg sync.WaitGroup
	wg.Add(6) // 2 incoming, 2 processing, 2 hooks

	go func() {
		for {
			ctx := context.Background()

			err := raydium.Start(ctx, wsUrl, raydiumProcessingCh)
			if err != nil {
				fmt.Printf("raydium.Start error, restarting: %v\n", err)
			}

			time.Sleep(3 * time.Second)

			fmt.Println("Raydium is restarting...")
		}
		// fmt.Println("Raydium out...")
		// wg.Done() // Signal completion of this goroutine
	}()

	go func() {
		for {
			ctx := context.Background()

			err := openbook.Start(ctx, wsUrl, openbookProcessingCh)
			if err != nil {
				fmt.Printf("openbook.Start error, restarting: %v\n", err)
				// if !strings.Contains(err.Error(), "EOF") {
				// 	break
				// }
			}

			time.Sleep(3 * time.Second)

			fmt.Println("Openbook is restarting...")
		}
		// fmt.Println("Openbook out...")
		// wg.Done() // Signal completion of this goroutine
	}()

	go func() {
		raydium.ProcessMessages(raydiumProcessingCh, raydiumHookCh)
		wg.Done()
	}()

	go func() {
		openbook.ProcessMessages(openbookProcessingCh, openbookHookCh)
		wg.Done()
	}()

	hooks.InitialiseDiscord()

	go func() {
		hooks.RaydiumDiscord(raydiumHookCh)
		wg.Done()
	}()

	go func() {
		hooks.OpenbookDiscord(openbookHookCh)
		wg.Done()
	}()

	wg.Wait()
}
