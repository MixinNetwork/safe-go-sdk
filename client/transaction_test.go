package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRPC(t *testing.T) {
	useClientServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transactions/transaction-id":
			if r.Method == http.MethodGet {
				writeJSON(t, w, Transaction{ID: "transaction-id", Chain: 1, State: "initial"})
				return
			}
			var request transactionRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			assert.Equal(t, int64(1), request.Chain)
			switch request.Action {
			case "approve":
				assert.Equal(t, "raw", request.Raw)
				writeJSON(t, w, Transaction{ID: "transaction-id", State: "approved"})
			case "revoke":
				assert.Equal(t, "signature", request.Signature)
				writeJSON(t, w, Transaction{ID: "transaction-id", State: "revoked"})
			default:
				t.Errorf("unexpected action %q", request.Action)
			}
		case "/transactions/missing":
			writeJSON(t, w, map[string]any{"error": 404})
		case "/transactions/error":
			writeJSON(t, w, map[string]any{"error": "denied"})
		case "/transactions/empty":
			writeJSON(t, w, map[string]any{})
		case "/transactions/malformed":
			_, _ = w.Write([]byte("not-json"))
		default:
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()
	tx, err := ReadTransaction(ctx, "transaction-id")
	require.NoError(t, err)
	assert.Equal(t, "transaction-id", tx.ID)
	for _, id := range []string{"missing", "empty"} {
		tx, err = ReadTransaction(ctx, id)
		require.NoError(t, err)
		assert.Nil(t, tx)
	}
	_, err = ReadTransaction(ctx, "error")
	assert.EqualError(t, err, "ReadTransaction error denied")
	_, err = ReadTransaction(ctx, "malformed")
	assert.Error(t, err)

	tx, err = ApproveTransaction(ctx, "transaction-id", 1, "raw")
	require.NoError(t, err)
	assert.Equal(t, "approved", tx.State)
	tx, err = ApproveTransaction(ctx, "missing", 1, "raw")
	require.NoError(t, err)
	assert.Nil(t, tx)
	_, err = ApproveTransaction(ctx, "error", 1, "raw")
	assert.EqualError(t, err, "ApproveTransaction error denied")
	_, err = ApproveTransaction(ctx, "malformed", 1, "raw")
	assert.Error(t, err)

	assert.NoError(t, RevokeTransaction(ctx, "transaction-id", 1, "signature"))
	assert.EqualError(t, RevokeTransaction(ctx, "error", 1, "signature"), "revoke error error denied")
	assert.Error(t, RevokeTransaction(ctx, "malformed", 1, "signature"))
}
