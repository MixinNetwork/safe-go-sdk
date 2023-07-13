package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/bitcoin"
)

type Account struct {
	ID      string           `json:"id"`
	Address string           `json:"address"`
	Chain   int64            `json:"chain"`
	Keys    []string         `json:"keys"`
	Outputs []*bitcoin.Input `json:"outputs"`
	Script  string           `json:"script"`
	State   string           `json:"state"`
	Error   any              `json:"error,omitempty"`
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
	Action    string `json:"action"`
	Address   string `json:"address"`
	Signature string `json:"signature"`
}

func ApproveAccount(ctx context.Context, id, address, signature string) (*Account, error) {
	req := accountRequest{
		Action:    "approve",
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

func CloseAccount(ctx context.Context, id, address, raw, hash string) (*Account, error) {
	req := map[string]string{
		"action":  "close",
		"address": address,
		"raw":     raw,
		"hash":    hash,
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
		return nil, fmt.Errorf("CloseAccount error %v", body.Error)
	}
	return &body, nil
}
