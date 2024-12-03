package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// Pool represents an AMM liquidity pool
type Pool struct {
	TokenA     string
	TokenB     string
	ReserveA   float64
	ReserveB   float64
	PoolPubKey solana.PublicKey
}

// Graph represents the exchange rate graph for arbitrage detection
type Graph struct {
	Vertices []string
	Edges    []Edge
	mu       sync.RWMutex
}

// Edge represents a directed edge in the exchange rate graph
type Edge struct {
	From   string
	To     string
	Weight float64 // Negative log of exchange rate
	Rate   float64
}

// RaydiumPoolState represents the state of a Raydium liquidity pool
type RaydiumPoolState struct {
	Status            uint64
	BaseDecimals      uint64
	QuoteDecimals     uint64
	LpDecimals        uint64
	BaseReserve       uint64
	QuoteReserve      uint64
	BaseTarget        uint64
	QuoteTarget       uint64
	BaseAmountPerRnd  uint64
	QuoteAmountPerRnd uint64
}

const (
	EPSILON = 1e-10 // Precision threshold for floating-point comparisons
)

// Helper function for comparing floating point numbers
func isSignificantlyDifferent(a, b float64) bool {
	return math.Abs(a-b) > EPSILON
}

func main() {

	ctx := context.Background()

	// Initialize Solana WebSocket client
	wsClient, err := ws.Connect(ctx, rpc.MainNetBeta_WS)
	if err != nil {
		log.Fatalf("Failed to connect to Solana WebSocket: %v", err)
	}
	defer wsClient.Close()

	// Initialize exchange rate graph
	graph := &Graph{
		Vertices: make([]string, 0),
		Edges:    make([]Edge, 0),
	}

	// Subscribe to account updates
	go monitorAccounts(ctx, wsClient, graph)

	// Start arbitrage detection loop
	detectArbitrage(graph)
}

// monitorAccounts subscribes to relevant pool account updates
func monitorAccounts(ctx context.Context, client *ws.Client, graph *Graph) {
	// Raydium pool accounts
	pools := map[string]struct {
		name       string
		baseToken  string
		quoteToken string
	}{
		"8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj": {
			name:       "USDC-SOL",
			baseToken:  "USDC",
			quoteToken: "SOL",
		},
		"2AXXcN6oN9bBT5owwmTH53C7QHUXvhLeu718Kqt8rvY2": {
			name:       "SOL-GRASS",
			baseToken:  "SOL",
			quoteToken: "GRASS",
		},
	}

	for poolPubKey, poolInfo := range pools {
		// Create a closure to capture pool info
		go func(pubKey string, info struct {
			name       string
			baseToken  string
			quoteToken string
		}) {
			poolAccount, err := solana.PublicKeyFromBase58(pubKey)
			if err != nil {
				log.Printf("Failed to parse pool public key %s: %v", pubKey, err)
				return
			}

			// Subscribe to account updates
			sub, err := client.AccountSubscribe(
				poolAccount,
				rpc.CommitmentConfirmed,
			)
			if err != nil {
				log.Printf("Failed to subscribe to account %s: %v", pubKey, err)
				return
			}
			defer sub.Unsubscribe()

			log.Printf("Successfully subscribed to Raydium pool %s (%s)", info.name, pubKey)

			// Start receiving updates
			for {
				select {
				case <-ctx.Done():
					return
				case update := <-sub.Response():
					if update.Value.Data == nil {
						continue
					}

					// Parse pool state
					poolState, err := parseRaydiumPoolState(update.Value.Data.GetBinary())
					if err != nil {
						log.Printf("Failed to parse pool state for %s: %v", info.name, err)
						continue
					}

					// Update graph with new exchange rates
					updateGraphWithPoolState(graph, poolState, info.baseToken, info.quoteToken)

					log.Printf("Pool Update (%s) - Base Reserve (%s): %d, Quote Reserve (%s): %d",
						info.name,
						info.baseToken,
						poolState.BaseReserve,
						info.quoteToken,
						poolState.QuoteReserve)
				}
			}
		}(poolPubKey, poolInfo)
	}

	// Keep the main goroutine running
	<-ctx.Done()
}

func parseRaydiumPoolState(data []byte) (*RaydiumPoolState, error) {
	if len(data) < 128 { // Minimum size for Raydium pool state
		return nil, fmt.Errorf("data too short for Raydium pool state")
	}

	state := &RaydiumPoolState{}

	// Raydium pool layout (8-byte aligned fields)
	state.Status = binary.LittleEndian.Uint64(data[0:8])
	state.BaseDecimals = binary.LittleEndian.Uint64(data[8:16])
	state.QuoteDecimals = binary.LittleEndian.Uint64(data[16:24])
	state.LpDecimals = binary.LittleEndian.Uint64(data[24:32])
	state.BaseReserve = binary.LittleEndian.Uint64(data[32:40])
	state.QuoteReserve = binary.LittleEndian.Uint64(data[40:48])
	state.BaseTarget = binary.LittleEndian.Uint64(data[48:56])
	state.QuoteTarget = binary.LittleEndian.Uint64(data[56:64])
	state.BaseAmountPerRnd = binary.LittleEndian.Uint64(data[64:72])
	state.QuoteAmountPerRnd = binary.LittleEndian.Uint64(data[72:80])

	return state, nil
}

func updateGraphWithPoolState(graph *Graph, state *RaydiumPoolState, baseToken, quoteToken string) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	// Use big.Float for precise calculations
	baseReserve := new(big.Float).SetUint64(state.BaseReserve)
	quoteReserve := new(big.Float).SetUint64(state.QuoteReserve)

	// Calculate rates with high precision
	baseToQuotePrice, _ := new(big.Float).Quo(quoteReserve, baseReserve).Float64()
	quoteToBasePrice, _ := new(big.Float).Quo(baseReserve, quoteReserve).Float64()

	// Apply fee with precision
	fee := 0.003
	baseToQuotePrice *= (1 - fee)
	quoteToBasePrice *= (1 - fee)

	// Convert to negative log with precision check
	baseToQuoteRate := -math.Log(baseToQuotePrice)
	quoteToBaseRate := -math.Log(quoteToBasePrice)

	// Check for invalid rates
	if math.IsInf(baseToQuoteRate, 0) || math.IsNaN(baseToQuoteRate) ||
		math.IsInf(quoteToBaseRate, 0) || math.IsNaN(quoteToBaseRate) {
		log.Printf("Warning: Invalid rate calculated for %s-%s pool", baseToken, quoteToken)
		return
	}

	log.Printf("Pool %s-%s: 1 %s = %.12f %s, 1 %s = %.12f %s",
		baseToken, quoteToken,
		baseToken, baseToQuotePrice, quoteToken,
		quoteToken, quoteToBasePrice, baseToken)

	// Update vertices if needed
	found := false
	for _, v := range graph.Vertices {
		if v == baseToken || v == quoteToken {
			found = true
			break
		}
	}

	if !found {
		graph.Vertices = append(graph.Vertices, baseToken, quoteToken)
		log.Printf("Added new vertices: %s, %s", baseToken, quoteToken)
	}

	// Update edges with precision handling
	graph.addEdge(baseToken, quoteToken, baseToQuotePrice)
	graph.addEdge(quoteToken, baseToken, quoteToBasePrice)
}

func (g *Graph) addEdge(from, to string, rate float64) {
	// For arbitrage detection:
	// If rate1 * rate2 * rate3 > 1 (profitable)
	// Then ln(rate1) + ln(rate2) + ln(rate3) > 0
	// And -ln(rate1) - ln(rate2) - ln(rate3) < 0 (negative cycle)
	weight := -math.Log(rate)
	g.Edges = append(g.Edges, Edge{
		From:   from,
		To:     to,
		Weight: weight,
		Rate:   rate,
	})
}

func bellmanFord(graph *Graph) [][]string {
	opportunities := make([][]string, 0)
	n := len(graph.Vertices)

	if n == 0 {
		return opportunities
	}

	// Try starting from each vertex
	for _, start := range graph.Vertices {
		dist := make(map[string]float64)
		prev := make(map[string]string)

		// Initialize all distances to infinity except start
		for _, v := range graph.Vertices {
			dist[v] = math.Inf(1)
		}
		dist[start] = 0

		// Relax edges |V| - 1 times
		for i := 0; i < n-1; i++ {
			for _, edge := range graph.Edges {
				if dist[edge.From] != math.Inf(1) {
					newDist := dist[edge.From] + edge.Weight
					if newDist < dist[edge.To] {
						dist[edge.To] = newDist
						prev[edge.To] = edge.From
					}
				}
			}
		}

		// Check for negative cycles (which indicate arbitrage opportunities)
		visited := make(map[string]bool)
		for _, edge := range graph.Edges {
			if dist[edge.From] != math.Inf(1) {
				newDist := dist[edge.From] + edge.Weight
				if newDist < dist[edge.To] {
					// Found a negative cycle (arbitrage opportunity)
					current := edge.From
					cycle := []string{current}
					visited[current] = true

					for {
						next := prev[current]
						if next == "" {
							break
						}
						if visited[next] {
							// Complete the cycle
							cycleStart := -1
							for i, v := range cycle {
								if v == next {
									cycleStart = i
									break
								}
							}
							if cycleStart != -1 {
								actualCycle := append(cycle[cycleStart:], next)

								// Calculate actual cycle profit
								amount := 1.0
								rates := make([]float64, 0)

								for i := 0; i < len(actualCycle)-1; i++ {
									from := actualCycle[i]
									to := actualCycle[i+1]

									// Find the direct exchange rate
									for _, e := range graph.Edges {
										if e.From == from && e.To == to {
											rate := math.Exp(-e.Weight) // Use exp(-weight) to get back original rate
											rates = append(rates, rate)
											amount *= rate
											break
										}
									}
								}

								profitPercent := (amount - 1.0) * 100

								log.Printf("Analyzing cycle: %v", actualCycle)
								log.Printf("Exchange rates: %v", rates)
								log.Printf("Final amount: %.12f (%.2f%%)", amount, profitPercent)

								// Only add to opportunities if profit is above threshold
								if amount > 1.0 { // Any profit is good for testing
									log.Printf("Found profitable cycle! Profit: %.2f%%", profitPercent)
									opportunities = append(opportunities, actualCycle)
								}
							}
							break
						}
						cycle = append(cycle, next)
						visited[next] = true
						current = next
					}
				}
			}
		}
	}

	return opportunities
}

func detectArbitrage(graph *Graph) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		graph.mu.RLock()
		if len(graph.Vertices) < 2 {
			log.Printf("Waiting for sufficient vertices... Current count: %d", len(graph.Vertices))
			graph.mu.RUnlock()
			continue
		}

		if len(graph.Edges) < 2 {
			log.Printf("Waiting for sufficient edges... Current count: %d", len(graph.Edges))
			graph.mu.RUnlock()
			continue
		}

		// Debug print current graph state
		log.Printf("Current Graph State - Vertices: %v", graph.Vertices)
		for _, edge := range graph.Edges {
			log.Printf("Edge: %s -> %s (Weight: %f)", edge.From, edge.To, edge.Weight)
		}

		opportunities := bellmanFord(graph)
		graph.mu.RUnlock()

		if len(opportunities) > 0 {
			log.Printf("Found %d arbitrage opportunities!", len(opportunities))
			printArbitrageOpportunities(opportunities)
		}
	}
}

// Printing arbitrage opportunities
func printArbitrageOpportunities(opportunities [][]string) {
	for i, path := range opportunities {
		if len(path) < 2 {
			continue
		}

		fmt.Printf("\n\n\n\n\nArbitrage Opportunity #%d:\n", i+1)
		fmt.Printf("Path: %s", path[0])
		for j := 1; j < len(path); j++ {
			fmt.Printf(" -> %s", path[j])
		}
		fmt.Printf("\n\n\n\n\n")
	}
}
