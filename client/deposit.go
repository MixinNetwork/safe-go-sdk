package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Deposit struct {
	Amount          string    `json:"amount"`
	AssetID         string    `json:"asset_id"`
	Chain           int64     `json:"chain"`
	Change          bool      `json:"change"`
	OutputIndex     int64     `json:"output_index"`
	Receiver        string    `json:"receiver"`
	TransactionHash string    `json:"transaction_hash"`
	Sender          string    `json:"sender"`
	SentHash        string    `json:"sent_hash"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func ReadDeposits(ctx context.Context, chain int64, offset int64) ([]*Deposit, error) {
	data, err := Request(ctx, "GET", fmt.Sprintf("/deposits?chain=%d&offset=%d", chain, offset), nil)
	if err != nil {
		return nil, err
	}

	var deposits []*Deposit
	err = json.Unmarshal(data, &deposits)
	if err != nil {
		return nil, err
	}
	return deposits, nil
}
