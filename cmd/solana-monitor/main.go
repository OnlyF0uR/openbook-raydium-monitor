package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/OnlyF0uR/solana-monitor/internal/hooks"
	"github.com/OnlyF0uR/solana-monitor/internal/hooks/discord_hook"
	"github.com/OnlyF0uR/solana-monitor/pkg/openbook"
	"github.com/OnlyF0uR/solana-monitor/pkg/raydium"
	"github.com/OnlyF0uR/solana-monitor/pkg/rpcs"
	"github.com/fatih/color"
	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
)

func main() {
	// Print a welcome message including the version, build date, and developer
	color.New(color.FgBlue).Println("============================================")
	color.New(color.FgCyan).Println("Welcome to Solana Monitor")
	color.New(color.FgCyan).Println("Version: 1.1.0")
	color.New(color.FgCyan).Println("Build Date: 2024-05-09")
	color.New(color.FgCyan).Println("Developer: OnlyF0uR (Discord: onlyspitfire)")
	color.New(color.FgCyan).Println("Reselling of this software is not allowed!")
	color.New(color.FgBlue).Println("============================================")

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Load RPCs
	rpcList := strings.Split(os.Getenv("SOLANA_RPC_URLS"), ";")
	rpcs.Initialise(rpcList)

	wsUrl := os.Getenv("SOLANA_WS_URL")

	// Channels for processing
	raydiumProcessingCh := make(chan solana.Signature)
	openbookProcessingCh := make(chan solana.Signature)
	// Channels for hooks
	raydiumHookCh := make(chan *raydium.RaydiumInfo)
	openbookHookCh := make(chan *openbook.OpenbookInfo)

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

	// Intialise the discord hooks
	discord_hook.Initialise()

	go func() {
		hooks.RunRaydiumHooks(raydiumHookCh)
		wg.Done()
	}()

	go func() {
		hooks.RunOpenbookHooks(openbookHookCh)
		wg.Done()
	}()

	wg.Wait()
}
