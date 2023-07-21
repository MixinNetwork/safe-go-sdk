package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type Recovery struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Chain    int64  `json:"chain"`
	Holder   string `json:"holder"`
	Observer string `json:"observer"`
	Hash     string `json:"hash"`
	Raw      string `json:"raw"`
	State    string `json:"state"`
	Error    any    `json:"error,omitempty"`
}

func ReadRecoveries(ctx context.Context) ([]*Recovery, error) {
	data, err := Request(ctx, "GET", "/recoveries", nil)
	if err != nil {
		return nil, err
	}
	var body []*Recovery
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func ReadRecovery(ctx context.Context, id string) (*Recovery, error) {
	data, err := Request(ctx, "GET", fmt.Sprintf("/recoveries/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var body Recovery
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}

	if body.Error != nil {
		if fmt.Sprint(body.Error) == "404" {
			return nil, nil
		}
		return nil, fmt.Errorf("ReadRecovery error %v", body.Error)
	}
	return &body, nil
}

type recoveryRequest struct {
	Raw  string `json:"raw"`
	Hash string `json:"hash"`
}

func SignRecovery(ctx context.Context, id, raw, hash string) (*Recovery, error) {
	req := recoveryRequest{
		Raw:  raw,
		Hash: hash,
	}
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	data, err := Request(ctx, "POST", fmt.Sprintf("/recoveries/%s", id), reqBuf)
	if err != nil {
		return nil, err
	}

	var body Recovery
	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}
	if body.Error != nil {
		if fmt.Sprint(body.Error) == "404" {
			return nil, nil
		}
		return nil, fmt.Errorf("SignRecovery error %v", body.Error)
	}
	return &body, nil
}
