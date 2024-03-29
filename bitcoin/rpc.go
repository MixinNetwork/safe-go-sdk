package bitcoin

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MixinNetwork/go-safe-sdk/common"
	"github.com/MixinNetwork/mixin/logger"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
)

type scriptPubKey struct {
	Hex             string   `json:"hex"`
	Type            string   `json:"type"`
	Address         string   `json:"address"`
	LegacyAddresses []string `json:"addresses"`
}

type rpcIn struct {
	Coinbase string `json:"coinbase"`
	TxId     string `json:"txid"`
	VOUT     int64  `json:"vout"`
}

type rpcOut struct {
	Value        float64       `json:"value"`
	N            int64         `json:"n"`
	ScriptPubKey *scriptPubKey `json:"scriptPubKey"`
}

type RPCTransaction struct {
	TxId      string    `json:"txid"`
	Vin       []*rpcIn  `json:"vin"`
	Vout      []*rpcOut `json:"vout"`
	BlockHash string    `json:"blockhash"`
	Hex       string    `json:"hex"`
}

type RPCBlock struct {
	Hash   string   `json:"hash"`
	Height uint64   `json:"height"`
	Tx     []string `json:"tx"`
	Time   int64    `json:"time"`
}

type RPCBlockWithTransactions struct {
	Hash   string            `json:"hash"`
	Height uint64            `json:"height"`
	Tx     []*RPCTransaction `json:"tx"`
}

func RPCGetTransactionOutput(chain byte, rpc, hash string, index int64) (*RPCTransaction, *Output, error) {
	cfg, err := common.NetConfig(chain)
	if err != nil {
		return nil, nil, err
	}
	tx, err := RPCGetTransaction(chain, rpc, hash)
	if err != nil {
		return nil, nil, err
	}
	if int64(len(tx.Vout)) <= index {
		return nil, nil, nil
	}
	out := tx.Vout[index]
	skt := out.ScriptPubKey.Type
	if skt != ScriptPubKeyTypeWitnessScriptHash && skt != ScriptPubKeyTypeWitnessKeyHash {
		return nil, nil, nil
	}
	if out.ScriptPubKey.Address == "" {
		return nil, nil, nil
	}

	satoshi := decimal.NewFromFloat(out.Value).Mul(decimal.NewFromFloat(ValueSatoshi))
	if !satoshi.IsInteger() || !satoshi.BigInt().IsInt64() {
		return nil, nil, nil
	}
	output := &Output{
		Address:  out.ScriptPubKey.Address,
		Satoshi:  satoshi.IntPart(),
		Coinbase: len(tx.Vin) == 0 && tx.Vin[0].Coinbase != "",
	}

	if tx.BlockHash == "" { // mempool
		output.Height = ^uint64(0)
	} else {
		block, err := RPCGetBlock(rpc, tx.BlockHash)
		if err != nil {
			return nil, nil, err
		}
		output.Height = block.Height
		output.Time = time.Unix(block.Time, 0)
	}

	rtb, err := hex.DecodeString(tx.Hex)
	if err != nil {
		return nil, nil, err
	}
	rtx, err := btcutil.NewTxFromBytes(rtb)
	if err != nil {
		return nil, nil, err
	}
	rtmx := rtx.MsgTx()
	if rtmx.TxHash().String() != hash {
		return nil, nil, nil
	}
	if len(rtmx.TxOut) != len(tx.Vout) {
		return nil, nil, nil
	}
	if rtmx.TxOut[index].Value != output.Satoshi {
		return nil, nil, nil
	}
	script, err := txscript.ParsePkScript(rtmx.TxOut[index].PkScript)
	if err != nil {
		return nil, nil, err
	}
	addr, err := script.Address(cfg)
	if err != nil {
		return nil, nil, err
	}
	if addr.EncodeAddress() != output.Address {
		return nil, nil, nil
	}

	return tx, output, nil
}

func RPCGetTransactionSender(chain byte, rpc string, tx *RPCTransaction) (string, error) {
	if tx.Vin[0].Coinbase != "" {
		return tx.Vin[0].Coinbase, nil
	}
	itx, err := RPCGetTransaction(chain, rpc, tx.Vin[0].TxId)
	if err != nil {
		return "", err
	}
	return itx.Vout[tx.Vin[0].VOUT].ScriptPubKey.Address, nil
}

func RPCGetTransaction(chain byte, rpc, hash string) (*RPCTransaction, error) {
	res, err := callBitcoinRPC(rpc, "getrawtransaction", []any{hash, 1})
	if err != nil {
		return nil, err
	}
	var tx RPCTransaction
	err = json.Unmarshal(res, &tx)
	if err != nil {
		return nil, err
	}
	fixLitecoinLegacyScriptPubKeyRPC(chain, &tx)
	return &tx, err
}

func RPCGetRawMempool(chain byte, rpc string) ([]*RPCTransaction, error) {
	res, err := callBitcoinRPC(rpc, "getrawmempool", []any{})
	if err != nil {
		return nil, err
	}
	var txs []string
	err = json.Unmarshal(res, &txs)
	if err != nil {
		return nil, err
	}

	var transactions []*RPCTransaction
	for _, id := range txs {
		tx, err := RPCGetTransaction(chain, rpc, id)
		if err != nil || tx == nil {
			logger.Printf("bitcoin.RPCGetRawMempool(%s) => %v %v", id, tx, err)
			continue
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func RPCGetBlockWithTransactions(chain byte, rpc, hash string) (*RPCBlockWithTransactions, error) {
	res, err := callBitcoinRPC(rpc, "getblock", []any{hash, 2})
	if err != nil {
		return nil, err
	}
	var b RPCBlockWithTransactions
	err = json.Unmarshal(res, &b)
	if err != nil {
		return nil, err
	}
	for _, tx := range b.Tx {
		fixLitecoinLegacyScriptPubKeyRPC(chain, tx)
		tx.BlockHash = hash
	}
	return &b, err
}

func RPCGetBlock(rpc, hash string) (*RPCBlock, error) {
	res, err := callBitcoinRPC(rpc, "getblock", []any{hash, 1})
	if err != nil {
		return nil, err
	}
	var b RPCBlock
	err = json.Unmarshal(res, &b)
	return &b, err
}

func RPCGetBlockHash(rpc string, num int64) (string, error) {
	res, err := callBitcoinRPC(rpc, "getblockhash", []any{num})
	if err != nil {
		return "", err
	}
	var hash string
	err = json.Unmarshal(res, &hash)
	return hash, err
}

func RPCGetBlockHeight(rpc string) (int64, error) {
	res, err := callBitcoinRPC(rpc, "getblockchaininfo", []any{})
	if err != nil {
		return 0, err
	}
	var info struct {
		Blocks int64 `json:"blocks"`
	}
	err = json.Unmarshal(res, &info)
	return info.Blocks, err
}

func RPCEstimateSmartFee(chain byte, rpc string) (int64, error) {
	res, err := callBitcoinRPC(rpc, "estimatesmartfee", []any{1})
	if err != nil {
		return 0, err
	}
	var fee struct {
		Rate float64 `json:"feerate"`
	}
	err = json.Unmarshal(res, &fee)
	if err != nil || fee.Rate <= 0 {
		return 0, fmt.Errorf("estimatesmartfee %f %v", fee.Rate, err)
	}
	fvb := int64(fee.Rate * 1.1 * ValueSatoshi / 1024)
	if fvb < 10 {
		fvb = 10
	}
	return fvb, nil
}

func RPCSendRawTransaction(rpc, raw string) (string, error) {
	res, err := callBitcoinRPC(rpc, "sendrawtransaction", []any{raw})
	if err != nil {
		return "", err
	}
	var hash string
	err = json.Unmarshal(res, &hash)
	return hash, err
}

// FIXME wait for litecoin core update to the latest rpc
func fixLitecoinLegacyScriptPubKeyRPC(chain byte, tx *RPCTransaction) {
	switch chain {
	case ChainLitecoin:
		for _, o := range tx.Vout {
			if len(o.ScriptPubKey.LegacyAddresses) != 1 {
				continue
			}
			o.ScriptPubKey.Address = o.ScriptPubKey.LegacyAddresses[0]
		}
	}
}

func callBitcoinRPC(rpc, method string, params []any) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	body, err := json.Marshal(map[string]any{
		"method":  method,
		"params":  params,
		"id":      time.Now().UnixNano(),
		"jsonrpc": "2.0",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", rpc, bytes.NewReader(body))
	if err != nil {
		return nil, buildRPCError(rpc, method, params, err)
	}

	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, buildRPCError(rpc, method, params, err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, buildRPCError(rpc, method, params, err)
	}
	var result struct {
		Data  any `json:"result"`
		Error any `json:"error"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("%v (%s)", buildRPCError(rpc, method, params, err), string(body))
	}
	if result.Error != nil {
		return nil, fmt.Errorf("%v (%s)", buildRPCError(rpc, method, params, err), string(body))
	}

	return json.Marshal(result.Data)
}

func buildRPCError(rpc, method string, params []any, err error) error {
	return fmt.Errorf("callBitcoinRPC(%s, %s, %v) => %v", rpc, method, params, err)
}
