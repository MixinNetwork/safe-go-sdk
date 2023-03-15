package safe

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	DefaultHost = "https://safe.mixin.dev"

	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{Timeout: 10 * time.Second}
}

func Request(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, DefaultHost+path, bytes.NewReader(body))
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
	return ioutil.ReadAll(resp.Body)
}
