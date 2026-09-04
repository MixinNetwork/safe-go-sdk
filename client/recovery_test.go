package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryRPC(t *testing.T) {
	useClientServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/recoveries":
			writeJSON(t, w, []*Recovery{{ID: "recovery-id", State: "initial", Chain: 2}})
		case "/recoveries/recovery-id":
			if r.Method == http.MethodPost {
				var request RecoveryRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				assert.Equal(t, "signature", request.Signature)
				assert.Equal(t, "approve", request.Action)
				writeJSON(t, w, Recovery{ID: "recovery-id", State: "signed"})
				return
			}
			writeJSON(t, w, Recovery{ID: "recovery-id", State: "initial"})
		case "/recoveries/missing":
			writeJSON(t, w, map[string]any{"error": 404})
		case "/recoveries/error":
			writeJSON(t, w, map[string]any{"error": "denied"})
		case "/recoveries/malformed", "/recoveries-malformed":
			_, _ = w.Write([]byte("not-json"))
		default:
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()
	recoveries, err := ReadRecoveries(ctx)
	require.NoError(t, err)
	require.Len(t, recoveries, 1)
	assert.Equal(t, "recovery-id", recoveries[0].ID)

	recovery, err := ReadRecovery(ctx, "recovery-id")
	require.NoError(t, err)
	assert.Equal(t, "initial", recovery.State)
	recovery, err = ReadRecovery(ctx, "missing")
	require.NoError(t, err)
	assert.Nil(t, recovery)
	_, err = ReadRecovery(ctx, "error")
	assert.EqualError(t, err, "ReadRecovery error denied")
	_, err = ReadRecovery(ctx, "malformed")
	assert.Error(t, err)

	recovery, err = SignRecovery(ctx, "recovery-id", RecoveryRequest{
		Signature: "signature",
		Raw:       "raw",
		Hash:      "hash",
		Id:        "request-id",
		Action:    "approve",
	})
	require.NoError(t, err)
	assert.Equal(t, "signed", recovery.State)
	recovery, err = SignRecovery(ctx, "missing", RecoveryRequest{})
	require.NoError(t, err)
	assert.Nil(t, recovery)
	_, err = SignRecovery(ctx, "error", RecoveryRequest{})
	assert.EqualError(t, err, "SignRecovery error denied")
	_, err = SignRecovery(ctx, "malformed", RecoveryRequest{})
	assert.Error(t, err)
}
