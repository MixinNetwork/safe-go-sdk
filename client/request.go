package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ProdHost = "https://observer.mixin.one"
	TestHost = "https://safe.mixin.dev"

	httpUri    string
	httpClient *http.Client
)

func init() {
	httpUri = ProdHost
	httpClient = &http.Client{Timeout: 1 * time.Minute}
}

func SetBaseUri(base string) {
	httpUri = base
}

func Request(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, httpUri+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("response status code %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
