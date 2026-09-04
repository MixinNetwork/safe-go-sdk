package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepositRPC(t *testing.T) {
	malformed := false
	useClientServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/deposits", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("chain"))
		assert.Equal(t, "17", r.URL.Query().Get("offset"))
		if malformed {
			_, _ = w.Write([]byte("not-json"))
			return
		}
		writeJSON(t, w, []*Deposit{{
			AssetID:         "asset-id",
			Amount:          "1.25",
			Chain:           2,
			TransactionHash: "transaction-hash",
			State:           "done",
		}})
	})

	deposits, err := ReadDeposits(context.Background(), 2, 17)
	require.NoError(t, err)
	require.Len(t, deposits, 1)
	assert.Equal(t, "asset-id", deposits[0].AssetID)
	assert.Equal(t, "1.25", deposits[0].Amount)

	malformed = true
	_, err = ReadDeposits(context.Background(), 2, 17)
	assert.Error(t, err)
}
