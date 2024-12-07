package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func getPricefromPool() {
	// Define token pairs (use actual mint addresses for the tokens)
	tokenIn := "So11111111111111111111111111111111111111112" // Example: SOL token mint address
	tokenOut := "USDCdQat9ykpo1V5vMe7Pfmq2tAdJXk6XkhkGVpHtF" // Example: USDC token mint address

	// Fetch price quote from Jupiter's Price API V2
	quote, err := FetchPriceQuote(tokenIn, tokenOut)
	if err != nil {
		log.Fatalf("Error fetching price quote: %v", err)
	}

	// Print the fetched price quote
	// Assuming quote is of type APIResponse
	// fmt.Printf("Price Quote for %s to %s:\n", tokenIn, tokenOut)
	// fmt.Printf("Price: %s\n", quote.Data.OutAmount) // Assuming 'OutAmount' is the correct field
	// fmt.Printf("Slippage: %s\n", quote.Data.Slippage)
	// fmt.Printf("Minimum Amount: %s\n", quote.Data.MinimumAmount)
	// fmt.Printf("Pool Count: %d\n", quote.Data.PoolCount)

	response, err := http.Get(jupiterAPIEndpoint)
	if err != nil {
		log.Fatalf("Error fetching price quote: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Raw API Response: %s\n", string(body)) // Print raw response

	// Parse the JSON response
	var responseData APIResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON response: %v", err)
	}

	// Output the price for each token
	for token, data := range responseData.Data {
		fmt.Printf("Token: %s, Price: %s\n", token, data.Price)
	}

	responseBody, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(responseBody)) // Print raw response for debugging
	fmt.Printf("Response: %+v\n", quote)

}
