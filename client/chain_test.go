package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainRPC(t *testing.T) {
	malformed := false
	unavailable := false
	useClientServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/chains", r.URL.Path)
		if unavailable {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if malformed {
			_, _ = w.Write([]byte("not-json"))
			return
		}
		writeJSON(t, w, []*Chain{
			{ID: "bitcoin", Chain: 1, Head: &Head{Height: 100}},
			{ID: "litecoin", Chain: 5, Sender: "observer"},
		})
	})

	chains, err := ReadChains(context.Background())
	require.NoError(t, err)
	require.Len(t, chains, 2)
	assert.Equal(t, int64(1), chains[0].Chain)
	assert.Equal(t, uint64(100), chains[0].Head.Height)
	assert.Equal(t, int64(5), chains[1].Chain)

	malformed = true
	_, err = ReadChains(context.Background())
	assert.Error(t, err)
	malformed = false
	unavailable = true
	_, err = ReadChains(context.Background())
	assert.EqualError(t, err, "response status code 500")
}

func TestRequestUsesContextAndValidatesRequests(t *testing.T) {
	useClientServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Request(ctx, http.MethodGet, "/context", nil)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = Request(context.Background(), "bad\nmethod", "/request", nil)
	assert.Error(t, err)
}
