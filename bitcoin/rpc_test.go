package bitcoin

import (
	"encoding/hex"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chainhash/v2"
	"github.com/btcsuite/btcd/wire/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bitcoinRPCRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
}

func bitcoinRPCServer(t *testing.T, responder func(bitcoinRPCRequest) (any, any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		var request bitcoinRPCRequest
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

func TestBitcoinRPCMethods(t *testing.T) {
	server := bitcoinRPCServer(t, func(request bitcoinRPCRequest) (any, any) {
		switch request.Method {
		case "getrawtransaction":
			id := request.Params[0].(string)
			if id == "missing" {
				return nil, map[string]any{"code": -5, "message": "not found"}
			}
			return map[string]any{
				"txid": id,
				"vin":  []any{map[string]any{"txid": "previous", "vout": 0}},
				"vout": []any{map[string]any{
					"value": 1.0,
					"n":     0,
					"scriptPubKey": map[string]any{
						"type":      ScriptPubKeyTypeWitnessKeyHash,
						"addresses": []string{"legacy-address"},
						"address":   "modern-address",
					},
				}},
			}, nil
		case "getrawmempool":
			return []string{"first", "missing", "second"}, nil
		case "getblock":
			verbosity := int(request.Params[1].(float64))
			if verbosity == 2 {
				return map[string]any{
					"hash":   "block-hash",
					"height": 88,
					"tx": []any{map[string]any{
						"txid": "inside",
						"vin":  []any{map[string]any{"coinbase": "abcd"}},
						"vout": []any{map[string]any{
							"scriptPubKey": map[string]any{"addresses": []string{"legacy-block-address"}},
						}},
					}},
				}, nil
			}
			return map[string]any{"hash": "block-hash", "height": 88, "tx": []string{"inside"}, "time": 1234}, nil
		case "getblockhash":
			return "block-at-height", nil
		case "getblockchaininfo":
			return map[string]any{"blocks": 999}, nil
		case "estimatesmartfee":
			return map[string]any{"feerate": 0.0002}, nil
		case "sendrawtransaction":
			return "sent-hash", nil
		default:
			return nil, map[string]any{"code": -32601, "message": "unknown method"}
		}
	})
	defer server.Close()

	tx, err := RPCGetTransaction(ChainLitecoin, server.URL, "first")
	require.NoError(t, err)
	assert.Equal(t, "first", tx.TxId)
	assert.Equal(t, "legacy-address", tx.Vout[0].ScriptPubKey.Address)

	mempool, err := RPCGetRawMempool(ChainBitcoin, server.URL)
	require.NoError(t, err)
	require.Len(t, mempool, 2)
	assert.Equal(t, "first", mempool[0].TxId)
	assert.Equal(t, "second", mempool[1].TxId)

	blockWithTransactions, err := RPCGetBlockWithTransactions(ChainLitecoin, server.URL, "block-hash")
	require.NoError(t, err)
	require.Len(t, blockWithTransactions.Tx, 1)
	assert.Equal(t, "block-hash", blockWithTransactions.Tx[0].BlockHash)
	assert.Equal(t, "legacy-block-address", blockWithTransactions.Tx[0].Vout[0].ScriptPubKey.Address)

	block, err := RPCGetBlock(server.URL, "block-hash")
	require.NoError(t, err)
	assert.Equal(t, uint64(88), block.Height)
	assert.Equal(t, int64(1234), block.Time)

	blockHash, err := RPCGetBlockHash(server.URL, 88)
	require.NoError(t, err)
	assert.Equal(t, "block-at-height", blockHash)
	height, err := RPCGetBlockHeight(server.URL)
	require.NoError(t, err)
	assert.Equal(t, int64(999), height)
	fee, err := RPCEstimateSmartFee(ChainBitcoin, server.URL)
	require.NoError(t, err)
	assert.Equal(t, int64(21), fee)
	sent, err := RPCSendRawTransaction(server.URL, "raw")
	require.NoError(t, err)
	assert.Equal(t, "sent-hash", sent)

	sender, err := RPCGetTransactionSender(ChainBitcoin, server.URL, &RPCTransaction{
		Vin: []*rpcIn{{Coinbase: "coinbase-data"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "coinbase-data", sender)
	sender, err = RPCGetTransactionSender(ChainBitcoin, server.URL, &RPCTransaction{
		Vin: []*rpcIn{{TxId: "first", VOUT: 0}},
	})
	require.NoError(t, err)
	assert.Equal(t, "modern-address", sender)
}

func TestRPCGetTransactionOutput(t *testing.T) {
	address, script := witnessAddress(t, ChainBitcoin, 0x33)
	msgTx := wire.NewMsgTx(2)
	msgTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{Hash: chainhash.Hash{1}, Index: 0},
		Sequence:         MaxTransactionSequence,
	})
	msgTx.AddTxOut(wire.NewTxOut(100_000, script))
	raw, err := MarshalWiredTransaction(msgTx, wire.BaseEncoding, ChainBitcoin)
	require.NoError(t, err)
	hash := msgTx.TxHash().String()
	coinbase := false

	server := bitcoinRPCServer(t, func(request bitcoinRPCRequest) (any, any) {
		switch request.Method {
		case "getrawtransaction":
			vin := []any{map[string]any{"txid": strings.Repeat("0", 64), "vout": 0}}
			if coinbase {
				vin = []any{map[string]any{"coinbase": "coinbase-data"}}
			}
			return map[string]any{
				"txid": hash,
				"vin":  vin,
				"vout": []any{map[string]any{
					"value": 0.001,
					"n":     0,
					"scriptPubKey": map[string]any{
						"type":    ScriptPubKeyTypeWitnessKeyHash,
						"address": address,
					},
				}},
				"blockhash": "block-hash",
				"hex":       hex.EncodeToString(raw),
			}, nil
		case "getblock":
			return map[string]any{"hash": "block-hash", "height": 321, "time": 1_700_000_000}, nil
		default:
			return nil, map[string]any{"message": "unexpected"}
		}
	})
	defer server.Close()

	tx, output, err := RPCGetTransactionOutput(ChainBitcoin, server.URL, hash, 0)
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.NotNil(t, output)
	assert.Equal(t, address, output.Address)
	assert.Equal(t, int64(100_000), output.Satoshi)
	assert.Equal(t, uint64(321), output.Height)
	assert.Equal(t, time.Unix(1_700_000_000, 0), output.Time)
	assert.False(t, output.Coinbase)
	coinbase = true
	_, output, err = RPCGetTransactionOutput(ChainBitcoin, server.URL, hash, 0)
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.True(t, output.Coinbase)

	tx, output, err = RPCGetTransactionOutput(ChainBitcoin, server.URL, hash, 1)
	require.NoError(t, err)
	assert.Nil(t, tx)
	assert.Nil(t, output)
	tx, output, err = RPCGetTransactionOutput(ChainBitcoin, server.URL, hash, -1)
	require.NoError(t, err)
	assert.Nil(t, tx)
	assert.Nil(t, output)
	_, _, err = RPCGetTransactionOutput(99, server.URL, hash, 0)
	assert.EqualError(t, err, "invalid chain 99")
}

func TestBitcoinRPCErrors(t *testing.T) {
	t.Run("rpc error", func(t *testing.T) {
		server := bitcoinRPCServer(t, func(bitcoinRPCRequest) (any, any) {
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
		_, err := callBitcoinRPC(server.URL, "method", nil)
		assert.ErrorContains(t, err, "not-json")
	})

	t.Run("connection error", func(t *testing.T) {
		_, err := callBitcoinRPC("://bad-url", "method", nil)
		assert.ErrorContains(t, err, "callBitcoinRPC")
	})

	t.Run("invalid fee", func(t *testing.T) {
		server := bitcoinRPCServer(t, func(bitcoinRPCRequest) (any, any) {
			return map[string]any{"feerate": 0}, nil
		})
		defer server.Close()
		_, err := RPCEstimateSmartFee(ChainBitcoin, server.URL)
		assert.ErrorContains(t, err, "estimatesmartfee")
	})

	t.Run("minimum fee", func(t *testing.T) {
		server := bitcoinRPCServer(t, func(bitcoinRPCRequest) (any, any) {
			return map[string]any{"feerate": math.SmallestNonzeroFloat64}, nil
		})
		defer server.Close()
		fee, err := RPCEstimateSmartFee(ChainBitcoin, server.URL)
		require.NoError(t, err)
		assert.Equal(t, int64(10), fee)
	})
}
