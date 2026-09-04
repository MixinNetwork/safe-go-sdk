package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountRPC(t *testing.T) {
	useClientServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/accounts/account-id":
			if r.Method == http.MethodGet {
				writeJSON(t, w, Account{ID: "account-id", Address: "safe-address", Chain: 1})
				return
			}
			var request map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			switch request["action"] {
			case "approve":
				assert.Equal(t, "holder-address", request["address"])
				assert.Equal(t, "signature", request["signature"])
			case "close":
				assert.Equal(t, "raw", request["raw"])
				assert.Equal(t, "hash", request["hash"])
			default:
				t.Errorf("unexpected action %q", request["action"])
			}
			writeJSON(t, w, Account{ID: "account-id", State: request["action"]})
		case "/accounts/account-id/inheritances":
			writeJSON(t, w, []*Inheritance{{LockId: "lock-id", Chain: 1, Status: "active"}})
		case "/accounts/missing", "/accounts/error-404":
			writeJSON(t, w, map[string]any{"error": 404})
		case "/accounts/error":
			writeJSON(t, w, map[string]any{"error": "denied"})
		case "/accounts/empty":
			writeJSON(t, w, map[string]any{})
		case "/accounts/malformed", "/accounts/malformed/inheritances":
			_, _ = w.Write([]byte("not-json"))
		case "/accounts/unavailable":
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()
	account, err := ReadAccount(ctx, "account-id")
	require.NoError(t, err)
	require.NotNil(t, account)
	assert.Equal(t, "account-id", account.ID)
	assert.Equal(t, int64(1), account.Chain)

	inheritances, err := ReadAccountInheritances(ctx, "account-id")
	require.NoError(t, err)
	require.Len(t, inheritances, 1)
	assert.Equal(t, "lock-id", inheritances[0].LockId)

	approved, err := ApproveAccount(ctx, "account-id", "holder-address", "signature")
	require.NoError(t, err)
	assert.Equal(t, "approve", approved.State)
	closed, err := CloseAccount(ctx, "account-id", "holder-address", "raw", "hash")
	require.NoError(t, err)
	assert.Equal(t, "close", closed.State)

	for _, id := range []string{"missing", "empty"} {
		account, err = ReadAccount(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, account)
	}
	_, err = ReadAccount(ctx, "error")
	assert.EqualError(t, err, "ReadAccount error denied")
	_, err = ReadAccount(ctx, "malformed")
	assert.Error(t, err)
	_, err = ReadAccount(ctx, "unavailable")
	assert.EqualError(t, err, "response status code 503")
	_, err = ReadAccountInheritances(ctx, "malformed")
	assert.Error(t, err)

	account, err = ApproveAccount(ctx, "error-404", "address", "signature")
	require.NoError(t, err)
	assert.Nil(t, account)
	_, err = ApproveAccount(ctx, "error", "address", "signature")
	assert.EqualError(t, err, "ApproveAccount error denied")
	account, err = CloseAccount(ctx, "error-404", "address", "raw", "hash")
	require.NoError(t, err)
	assert.Nil(t, account)
	_, err = CloseAccount(ctx, "error", "address", "raw", "hash")
	assert.EqualError(t, err, "CloseAccount error denied")
}
