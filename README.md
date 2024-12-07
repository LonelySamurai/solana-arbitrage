# Solana Arbitrage Detector

A real-time arbitrage detection bot for Solana's Raydium DEX that monitors liquidity pools and identifies profitable trading opportunities using the Bellman-Ford algorithm.

## Features

- Real-time monitoring of Raydium liquidity pools via WebSocket connection
- Automatic detection of arbitrage opportunities across trading pairs
- Efficient negative cycle detection using Bellman-Ford algorithm
- Detailed logging of pool states and potential profit opportunities

## Prerequisites

- Go 1.x
- Solana RPC endpoint access

## Dependencies

```
github.com/gagliardetto/solana-go
```

## Running the Program

1. Clone the repository
2. Install dependencies:
   ```
   go mod download
   ```
3. Run the program:
   ```
   go run main.go
   ```

The program will connect to Solana's mainnet, monitor specified Raydium pools in the code (Change the address to your desrieed pool addresses), and automatically detect and log any arbitrage opportunities as they arise.

## Current Monitored Pools

- USDC-SOL
- SOL-GRASS

Last updated: 2024-12-07
