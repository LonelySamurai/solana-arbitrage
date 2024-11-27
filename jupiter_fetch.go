package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type JupiterPool struct {
	ID      string  `json:"id"`
	TokenA  string  `json:"tokenA"`
	TokenB  string  `json:"tokenB"`
	PriceAB float64 `json:"priceAB"`
	PriceBA float64 `json:"priceBA"`
}

// PriceQuote represents the structure of a price quote from Jupiter's Price API V2
type PriceQuote struct {
	Slippage      string `json:"slippage"`
	MinimumAmount string `json:"minimumAmount"`
	Price         string `json:"price"`
	TokenIn       string `json:"tokenIn"`
	TokenOut      string `json:"tokenOut"`
	PoolCount     int    `json:"poolCount"`
}

type PriceInfo struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Price string `json:"price"`
}

type QuoteResponse struct {
	Data      map[string]PriceInfo `json:"data"` // The map where each key is a token ID
	TimeTaken float64              `json:"timeTaken"`
}

// type QuoteResponse struct {
// 	Data struct {
// 		InAmount      string `json:"inAmount"`
// 		OutAmount     string `json:"outAmount"`
// 		Slippage      string `json:"slippage"`
// 		MinimumAmount string `json:"minimumAmount"`
// 		PoolCount     int    `json:"poolCount"`
// 	} `json:"data"`
// }

const jupiterAPIEndpoint = "https://api.jup.ag/price/v2?ids=JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN,So11111111111111111111111111111111111111112"

// FetchPriceQuote fetches price quotes from Jupiter's Price API V2 for a specific token pair.
func FetchPriceQuote(tokenIn, tokenOut string) (*QuoteResponse, error) {
	// Construct the request URL with dynamic token mint addresses
	url := fmt.Sprintf("https://api.jup.ag/price/v2?ids=%s,%s", tokenIn, tokenOut)

	// Make the GET request to the Jupiter API
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %v", resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Log raw response for debugging
	fmt.Printf("Raw API Response: %s\n", string(body))

	// Parse the JSON response into QuoteResponse struct
	var quoteResp QuoteResponse
	if err := json.Unmarshal(body, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to decode price quote: %v", err)
	}

	// Return the structured quote response
	return &quoteResp, nil
}

// Construct the request URL with query parameters
// url := fmt.Sprintf("%s?inputMint=%s&outputMint=%s", jupiterAPIEndpoint, tokenIn, tokenOut)

// // Create a HTTP client with a timeout
// client := &http.Client{Timeout: 10 * time.Second}

// // Send the request to the Jupiter API
// resp, err := client.Get(url)
// if err != nil {
// 	return nil, fmt.Errorf("failed to send request: %v", err)
// }
// defer resp.Body.Close()

// // Check for non-2xx response codes
// if resp.StatusCode != http.StatusOK {
// 	return nil, fmt.Errorf("HTTP error: %s", resp.Status)
// }

// // Decode the response body into the PriceQuote struct
// var quote PriceQuote
// if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
// 	return nil, fmt.Errorf("failed to decode response: %v", err)
// }

// return &quote, nil
//}
