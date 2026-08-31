package client

import (
	"context"
	"encoding/json"
	"time"
)

type AccountantOutputs struct {
	Count   int    `json:"count"`
	Satoshi uint64 `json:"satoshi"`
}

type Accountant struct {
	Outputs AccountantOutputs `json:"outputs"`
}

type Head struct {
	CreatedAt time.Time `json:"created_at"`
	Fee       int64     `json:"fee"`
	Hash      string    `json:"hash"`
	Height    uint64    `json:"height"`
	ID        string    `json:"id"`
}

type ChainOperationParams struct {
	Asset string `json:"asset"`
	Price string `json:"price"`
}

type ChainTransactionParams struct {
	Minimum string `json:"minimum"`
}

type ChainParams struct {
	Operation   ChainOperationParams   `json:"operation"`
	Transaction ChainTransactionParams `json:"transaction"`
}

type Chain struct {
	ID         string      `json:"id"`
	Chain      int64       `json:"chain"`
	Head       *Head       `json:"head"`
	Params     ChainParams `json:"params"`
	Accountant Accountant  `json:"accountant,omitempty"` // For bitcoin, litecoin
	Sender     string      `json:"sender,omitempty"`     // For evm chains
}

func ReadChains(ctx context.Context) ([]*Chain, error) {
	data, err := Request(ctx, "GET", "/chains", nil)
	if err != nil {
		return nil, err
	}

	var body []*Chain
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
