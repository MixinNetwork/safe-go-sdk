package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ProdHost       = "https://observer.mixin.one"
	TestHost       = "https://safe.mixin.dev"
	keyEnvironment = "env"

	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{Timeout: 10 * time.Second}
}

func getHost(ctx context.Context) string {
	host := ProdHost
	env, _ := ctx.Value(keyEnvironment).(string)
	if env != "prod" {
		host = TestHost
	}
	return host
}

func Request(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	host := getHost(ctx)
	req, err := http.NewRequest(method, host+path, bytes.NewReader(body))
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
