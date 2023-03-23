package inquiries

import (
	"context"
	"encoding/json"
)

type Head struct {
	Fee    int64  `json:"fee"`
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
	ID     string `json:"id"`
}

type Chain struct {
	ID    string `json:"id"`
	Chain int64  `json:"chain"`
	Head  *Head  `json:"head"`
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
