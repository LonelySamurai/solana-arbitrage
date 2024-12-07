package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// ConnectWebSocket initializes a WebSocket connection to Solana
func ConnectWebSocket(url string) *websocket.Conn {
	// Connect to Solana WebSocket endpoint
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Failed to connect to Solana WebSocket: %v", err)
	}
	log.Println("Connected to Solana WebSocket")
	return conn
}

// SubscribeToAccount subscribes to updates for a specific account
func SubscribeToAccount(accountPubKey string) {
	url := "wss://api.mainnet-beta.solana.com"
	conn := ConnectWebSocket(url)
	defer conn.Close()

	// Prepare the subscription payload
	subscribePayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "accountSubscribe",
		"params": []interface{}{
			accountPubKey,
			map[string]string{
				"commitment": "confirmed",
			},
		},
	}

	// Send subscription request
	err := conn.WriteJSON(subscribePayload)
	if err != nil {
		log.Fatalf("Failed to send subscription request: %v", err)
	}

	// Listen for messages
	fmt.Println("Listening for account updates...")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("Error reading message: %v", err)
		}
		fmt.Printf("Received: %s\n", message)
	}
}
