package inquiries

import (
	"context"
	"encoding/json"
	"fmt"
)

type Output struct {
	TransactionHash string `json:"transaction_hash"`
	Satoshi         int64  `json:"satoshi"`
	OutputIndex     uint32 `json:"output_index"`
}

type Accountant struct {
	Address string    `json:"address"`
	Outputs []*Output `json:"outputs"`
}

type Account struct {
	ID         string      `json:"id"`
	Accountant *Accountant `json:"accountant"`
	Address    string      `json:"address"`
	Chain      int64       `json:"chain"`
	Outputs    []*Output   `json:"outputs"`
	Script     string      `json:"script"`
	Status     string      `json:"status"`
	Error      any         `json:"error,omitempty"`
}

func ReadAccount(ctx context.Context, id string) (*Account, error) {
	data, err := Request(ctx, "GET", fmt.Sprintf("/accounts/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var body Account
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	if body.Error != nil {
		if fmt.Sprint(body.Error) == "404" {
			return nil, nil
		}
		return nil, fmt.Errorf("ReadAccount error %v", body.Error)
	}
	if body.ID == "" {
		return nil, nil
	}
	return &body, nil
}

type accountRequest struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
}

func ApproveAccount(ctx context.Context, id, address, signature string) (*Account, error) {
	req := accountRequest{
		Address:   address,
		Signature: signature,
	}
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err := Request(ctx, "POST", fmt.Sprintf("/accounts/%s", id), reqBuf)
	if err != nil {
		return nil, err
	}

	var body Account
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	if body.Error != nil {
		if fmt.Sprint(body.Error) == "404" {
			return nil, nil
		}
		return nil, fmt.Errorf("ApproveAccount error %v", body.Error)
	}
	return &body, nil
}