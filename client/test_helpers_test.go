package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func useClientServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	oldURI := httpUri
	oldClient := httpClient
	server := httptest.NewServer(handler)
	SetBaseUri(server.URL)
	httpClient = server.Client()
	t.Cleanup(func() {
		server.Close()
		httpUri = oldURI
		httpClient = oldClient
	})
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(value))
}
