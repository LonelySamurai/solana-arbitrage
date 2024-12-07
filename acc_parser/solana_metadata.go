package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func solana_metadata() {
	// Create an RPC client instance (Solana mainnet)
	client := rpc.New(rpc.MainNetBeta.RPC)

	// Public key of the account
	accountPubKey := solana.MustPublicKeyFromBase58("9xqnnfeonbsEGSPgF5Wd7bf9RqXy4KP22bdaGmZbHGwp")

	// Fetch account information
	accountInfo, err := client.GetAccountInfo(context.Background(), accountPubKey)
	if err != nil {
		log.Fatalf("Failed to fetch account info: %v", err)
	}

	// Decode the account's raw data (assuming it's a token account)
	var parsedTokenAccount TokenAccount
	if err := parsedTokenAccount.Decode(accountInfo.Value.Data.GetBinary()); err != nil {
		log.Fatalf("Failed to decode account data: %v", err)
	}

	// Print the parsed information
	fmt.Printf("Parsed Token Account Data: %+v\n\n", parsedTokenAccount)

	// Print the token balance (UIAmount)
	fmt.Printf("Token Account Balance: %.6f\n", parsedTokenAccount.TokenAmount.UIAmount)

	// Fetch the native SOL balance (in lamports, then converted to SOL)
	commitment := rpc.CommitmentFinalized // Specify the commitment level
	balanceResult, err := client.GetBalance(context.Background(), accountPubKey, commitment)
	if err != nil {
		log.Fatalf("Failed to fetch SOL balance: %v", err)
	}

	// Access the balance value (which is in lamports)
	balance := balanceResult.Value

	// Print the SOL balance (converted from lamports to SOL)
	fmt.Printf("Native SOL Balance: %.6f SOL\n", float64(balance)/1e9) // Converting from lamports to SOL
}
