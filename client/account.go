package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Output struct {
	TransactionHash string `json:"transaction_hash"`
	Satoshi         int64  `json:"satoshi"`
	OutputIndex     uint32 `json:"output_index"`
	Script          string `json:"script"`
	Sequence        uint32 `json:"sequence"`
}

type AssetBalance struct {
	AssetAddress string `json:"asset_address"`
	Amount       string `json:"amount"`
	SafeAssetId  string `json:"safe_asset_id"`
}

type Account struct {
	ID              string                  `json:"id"`
	Address         string                  `json:"address"`
	Chain           int64                   `json:"chain"`
	Keys            []string                `json:"keys"`
	Outputs         []Output                `json:"outputs"`          // For bitcoin, litecoin: unspent outputs,
	Pendings        []Output                `json:"pendings"`         // For bitcoin, litecoin: signed outptus,
	Changes         []Output                `json:"changes"`          // For bitcoin, litecoin: unreceived changes
	Balances        map[string]AssetBalance `json:"balances"`         // For evm chains: unspent balances,
	PendingBalances map[string]AssetBalance `json:"pending_balances"` // For evm chains: signed balances,
	Nonce           int64                   `json:"nonce"`            // For evm chains
	Script          string                  `json:"script"`
	State           string                  `json:"state"`
	Migrated        bool                    `json:"migrated"`
	SafeAssetId     string                  `json:"safe_asset_id"`
	Error           any                     `json:"error,omitempty"`
}

type Inheritance struct {
	LockId    string    `json:"lock_id"`
	RequestId string    `json:"request_id"`
	Hash      string    `json:"hash"`
	Holder    string    `json:"holder"`
	Address   string    `json:"address"`
	Chain     int64     `json:"chain"`
	Duration  int64     `json:"duration"` // lock for hours
	Status    string    `json:"state"`    // initial | active | revoked
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func ReadAccountInheritances(ctx context.Context, id string) ([]*Inheritance, error) {
	data, err := Request(ctx, "GET", fmt.Sprintf("/accounts/%s/inheritances", id), nil)
	if err != nil {
		return nil, err
	}

	var inheritances []*Inheritance
	err = json.Unmarshal(data, &inheritances)
	if err != nil {
		return nil, err
	}
	return inheritances, nil
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
