package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gorilla/websocket"
)

type TokenAccount struct {
	IsNative    bool   `json:"isNative"`
	Mint        string `json:"mint"`
	Owner       string `json:"owner"`
	State       string `json:"state"`
	TokenAmount struct {
		Amount         string  `json:"amount"`
		Decimals       int     `json:"decimals"`
		UIAmount       float64 `json:"uiAmount"`
		UIAmountString string  `json:"uiAmountString"`
	} `json:"tokenAmount"`
}

type ApiPoolInfoV4 struct {
	ID            string `json:"id"`
	BaseMint      string `json:"baseMint"`
	QuoteMint     string `json:"quoteMint"`
	LpMint        string `json:"lpMint"`
	BaseVault     string `json:"baseVault"`
	QuoteVault    string `json:"quoteVault"`
	OpenOrders    string `json:"openOrders"`
	Authority     string `json:"authority"`
	Version       uint8  `json:"version"`
	BaseDecimals  uint8  `json:"baseDecimals"`
	QuoteDecimals uint8  `json:"quoteDecimals"`
}

type PriceData struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Price string `json:"price"`
}

type SubscriptionMessage struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"` // Params can vary based on method
	Id      int         `json:"id"`
}

// AccountChangeParams represents the parameters for subscribing to account changes
type AccountChangeParams struct {
	Account string `json:"account"`
}

type APIResponse struct {
	Data      map[string]PriceData `json:"data"`
	TimeTaken float64              `json:"timeTaken"`
}

// ///////////// Decode  raw account data into the TokenAccount struct/////////////
func (ta *TokenAccount) Decode(data []byte) error {
	// Decode Mint (32 bytes)
	ta.Mint = solana.PublicKeyFromBytes(data[:32]).String()

	// Decode Owner (32 bytes)
	ta.Owner = solana.PublicKeyFromBytes(data[32:64]).String()

	// Decode Amount (8 bytes)
	amountBytes := data[64:72]
	amount := binary.LittleEndian.Uint64(amountBytes)
	ta.TokenAmount.Amount = fmt.Sprintf("%d", amount)

	// Decode State (1 byte)
	stateByte := data[108]
	if stateByte == 1 {
		ta.State = "initialized"
	} else {
		ta.State = "uninitialized"
	}

	// Decode Decimals (typically 6 for USDC and SPL tokens)
	ta.TokenAmount.Decimals = 6 // Hardcoded for simplicity; you might fetch dynamically based on token

	// Calculate the UIAmount
	ta.TokenAmount.UIAmount = float64(amount) / float64(1e6) // Assuming 6 decimals
	ta.TokenAmount.UIAmountString = fmt.Sprintf("%.6f", ta.TokenAmount.UIAmount)

	return nil

}

const wsURL = "wss://api.mainnet-beta.solana.com"                  // Solana WebSocket RPC URL
const poolAccount = "8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj" //Raydium SOL-USDC Pool ID

type WebSocketMessage struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		SubscriptionId string `json:"subscriptionId"`
		Result         struct {
			Value struct {
				ProgramId string `json:"programId"`
				Accounts  []struct {
					Pubkey string `json:"pubkey"`
					Data   string `json:"data"`
				} `json:"accounts"`
			} `json:"value"`
		} `json:"result"`
	} `json:"params"`
}

// //////////////
// func decodeRaydiumPoolData(data []byte) (ApiPoolInfoV4, error) {
// 	var poolInfo ApiPoolInfoV4

// 	// Here we use the binary decoding function for Raydium pool data
// 	// Assuming the data is correctly formatted, decode the binary blob into ApiPoolInfoV4 struct
// 	err := encoding.Decode(data, &poolInfo)
// 	if err != nil {
// 		return ApiPoolInfoV4{}, fmt.Errorf("failed to decode Raydium pool data: %w", err)
// 	}
// 	return poolInfo, nil
// }

// ///////////////////////////////////////////////////////////////////////
func main() {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	//  subscription message for the specific Raydium pool account
	subscribeMessage := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "accountSubscribe",
		"params":  []string{poolAccount},
		"id":      1,
	}

	//  subscription message to the WebSocket server
	if err := conn.WriteJSON(subscribeMessage); err != nil {
		log.Fatalf("Failed to send subscription message: %v", err)
	}

	// listening for messages
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		
		fmt.Printf("Received message: %+v\n", msg)

		
	}
}
