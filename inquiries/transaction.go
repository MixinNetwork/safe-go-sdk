package inquiries

import (
	"context"
	"encoding/json"
	"fmt"
)

type Transaction struct {
	ID      string   `json:"id"`
	Chain   int64    `json:"chain"`
	Fee     string   `json:"fee"`
	Hash    string   `json:"hash"`
	Raw     string   `json:"raw"`
	Signers []string `json:"signers"`
	Error   any      `json:"error,omitempty"`
}

func ReadTransaction(ctx context.Context, id string) (*Transaction, error) {
	data, err := Request(ctx, "GET", fmt.Sprintf("/transactions/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var body Transaction
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	if body.Error != nil {
		if fmt.Sprint(body.Error) == "404" {
			return nil, nil
		}
		return nil, fmt.Errorf("ReadTransaction error %v", body.Error)
	}
	if body.ID == "" {
		return nil, nil
	}
	return &body, nil
}

type transactionRequest struct {
	Action    string `json:"action"`
	Chain     int64  `json:"chain"`
	Raw       string `json:"raw"`
	Signature string `json:"signature"`
}

func ApproveTransaction(ctx context.Context, id string, chain int64, raw, signature string) (*Transaction, error) {
	req := transactionRequest{
		Action:    "approve",
		Chain:     chain,
		Raw:       raw,
		Signature: signature,
	}
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err := Request(ctx, "POST", fmt.Sprintf("/transactions/%s", id), reqBuf)
	if err != nil {
		return nil, err
	}

	var body Transaction
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	if body.Error != nil {
		if fmt.Sprint(body.Error) == "404" {
			return nil, nil
		}
		return nil, fmt.Errorf("ApproveTransaction error %v", body.Error)
	}
	return &body, nil
}
