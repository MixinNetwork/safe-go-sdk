package ethereum

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ethereumRPCRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

func ethereumRPCServer(t *testing.T, responder func(ethereumRPCRequest) (any, any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var request ethereumRPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		result, rpcError := responder(request)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  result,
			"error":   rpcError,
		}))
	}))
}

func TestEthereumRPCMethods(t *testing.T) {
	server := ethereumRPCServer(t, func(request ethereumRPCRequest) (any, any) {
		switch request.Method {
		case "eth_getBlockByHash":
			fullTransactions := request.Params[1].(bool)
			if fullTransactions {
				return map[string]any{
					"hash":   "0xblock",
					"number": "0x2a",
					"transactions": []any{map[string]any{
						"hash":        "0xtx",
						"blockNumber": "0x2a",
					}},
				}, nil
			}
			return map[string]any{
				"hash":         "0xblock",
				"number":       "0x2a",
				"transactions": []string{"0xtx"},
				"timestamp":    "0x65",
			}, nil
		case "eth_blockNumber":
			return "0x2a", nil
		case "eth_getBlockByNumber":
			require.Len(t, request.Params, 2)
			assert.Equal(t, "0x0", request.Params[0])
			assert.Equal(t, false, request.Params[1])
			return map[string]any{"hash": "0xgenesis", "number": "0x0"}, nil
		case "eth_gasPrice":
			return "0x64", nil
		case "eth_getTransactionByHash":
			return map[string]any{
				"hash":        request.Params[0],
				"blockHash":   "0xblock",
				"blockNumber": "0x2a",
				"from":        testSafeAddress,
				"to":          testDestination,
			}, nil
		case "eth_getBalance":
			return "0xff", nil
		case "debug_traceTransaction":
			assert.Equal(t, "0xabc", request.Params[0])
			return map[string]any{"type": "CALL", "from": testSafeAddress, "to": testDestination, "gasUsed": "0x10"}, nil
		case "debug_traceBlockByHash":
			return []any{map[string]any{"result": map[string]any{"type": "CALL", "from": testSafeAddress, "to": testDestination}}}, nil
		default:
			return nil, map[string]any{"code": -32601, "message": "unknown method"}
		}
	})
	defer server.Close()

	block, err := RPCGetBlock(server.URL, "0xblock")
	require.NoError(t, err)
	assert.Equal(t, uint64(42), block.Height)
	assert.Equal(t, time.Unix(101, 0), block.Time)
	assert.Equal(t, []string{"0xtx"}, block.Tx)

	height, err := RPCGetBlockHeight(server.URL)
	require.NoError(t, err)
	assert.Equal(t, int64(42), height)
	blockHash, err := RPCGetBlockHash(server.URL, 0)
	require.NoError(t, err)
	assert.Equal(t, "0xgenesis", blockHash)
	_, err = RPCGetBlockHash(server.URL, -1)
	assert.EqualError(t, err, "invalid block height -1")

	fullBlock, err := RPCGetBlockWithTransactions(server.URL, "0xblock")
	require.NoError(t, err)
	assert.Equal(t, uint64(42), fullBlock.Height)
	require.Len(t, fullBlock.Tx, 1)
	assert.Equal(t, "0xblock", fullBlock.Tx[0].BlockHash)

	gasPrice, err := RPCGetGasPrice(server.URL)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(100), gasPrice)
	tx, err := RPCGetTransactionByHash(server.URL, "0xtx")
	require.NoError(t, err)
	assert.Equal(t, uint64(42), tx.BlockHeight)
	assert.Equal(t, "0xblock", tx.BlockHash)

	balance, err := RPCGetAddressBalance(server.URL, "0xtx", testDestination)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(255), balance)
	balance, err = RPCGetAddressBalanceAtBlock(server.URL, "0xblock", testDestination)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(255), balance)

	trace, err := RPCDebugTraceTransactionByHash(server.URL, "abc")
	require.NoError(t, err)
	assert.Equal(t, "CALL", trace.Type)
	assert.Equal(t, "0x10", trace.GasUsed)
	traces, err := RPCDebugTraceBlockByHash(server.URL, "0xblock")
	require.NoError(t, err)
	require.Len(t, traces, 1)
	assert.Equal(t, "CALL", traces[0].Result.Type)
}

func TestEthereumNumberToUint64(t *testing.T) {
	for input, want := range map[string]uint64{
		"0x0":                0,
		"0x2a":               42,
		"0xffffffffffffffff": ^uint64(0),
	} {
		got, err := ethereumNumberToUint64(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}

	for _, input := range []string{"", "1", "0x", "0xzz", "0x10000000000000000"} {
		t.Run(input, func(t *testing.T) {
			_, err := ethereumNumberToUint64(input)
			assert.Error(t, err)
		})
	}
}

func TestEthereumNumberToBigInt(t *testing.T) {
	value, err := ethereumNumberToBigInt("0x10000000000000000")
	require.NoError(t, err)
	assert.Equal(t, "18446744073709551616", value.String())

	for _, input := range []string{"", "1", "bad", "0x", "0xzz"} {
		t.Run(input, func(t *testing.T) {
			_, err := ethereumNumberToBigInt(input)
			assert.Error(t, err)
		})
	}
}

func TestEthereumRPCErrors(t *testing.T) {
	t.Run("rpc error", func(t *testing.T) {
		server := ethereumRPCServer(t, func(ethereumRPCRequest) (any, any) {
			return nil, map[string]any{"code": -1, "message": "boom"}
		})
		defer server.Close()
		_, err := RPCGetBlockHeight(server.URL)
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("invalid response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()
		_, err := callEthereumRPC(server.URL, "method", nil)
		assert.ErrorContains(t, err, "not-json")
	})

	t.Run("connection error", func(t *testing.T) {
		_, err := callEthereumRPCUntilSufficient("://bad-url", "method", nil)
		assert.ErrorContains(t, err, "callEthereumRPC")
	})

	t.Run("invalid gas prices", func(t *testing.T) {
		responses := []any{"100", "0xzz"}
		for _, response := range responses {
			server := ethereumRPCServer(t, func(ethereumRPCRequest) (any, any) { return response, nil })
			_, err := RPCGetGasPrice(server.URL)
			assert.ErrorContains(t, err, "invalid hex")
			server.Close()
		}
	})
}
