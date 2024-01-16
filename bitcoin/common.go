package bitcoin

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	commonSafe "github.com/MixinNetwork/go-safe-sdk/common"
	"github.com/MixinNetwork/mixin/common"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/shopspring/decimal"
)

const (
	ChainBitcoin  = 1
	ChainLitecoin = 5

	ValuePrecision = 8
	ValueSatoshi   = 100000000

	TimeLockMinimum = time.Hour * 1
	TimeLockMaximum = time.Hour * 24 * 365

	ScriptPubKeyTypeWitnessKeyHash    = "witness_v0_keyhash"
	ScriptPubKeyTypeWitnessScriptHash = "witness_v0_scripthash"
	SigHashType                       = txscript.SigHashAll | txscript.SigHashAnyOneCanPay

	InputTypeP2WPKHAccoutant             = 1
	InputTypeP2WSHMultisigHolderSigner   = 2
	InputTypeP2WSHMultisigObserverSigner = 3

	MaxTransactionSequence = 0xffffffff
	MaxStandardTxWeight    = 300000

	TransactionConfirmations = 1
)

func ParseSatoshi(amount string) (int64, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, err
	}
	amt = amt.Mul(decimal.New(1, ValuePrecision))
	if !amt.IsInteger() {
		return 0, fmt.Errorf("invalid amount %s", amount)
	}
	if !amt.BigInt().IsInt64() {
		return 0, fmt.Errorf("invalid amount %s", amount)
	}
	return amt.BigInt().Int64(), nil
}

func ParseAddress(addr string, chain byte) ([]byte, error) {
	switch chain {
	case ChainBitcoin:
		err := VerifyAddress(addr, CoinBitcoin)
		if err != nil {
			return nil, fmt.Errorf("bitcoin.VerifyAddress(%s) => %v", addr, err)
		}
	case ChainLitecoin:
		err := VerifyAddress(addr, CoinLitecoin)
		if err != nil {
			return nil, fmt.Errorf("litecoin.VerifyAddress(%s) => %v", addr, err)
		}
	default:
		return nil, fmt.Errorf("ParseAddress(%s, %d)", addr, chain)
	}
	cfg, _ := commonSafe.NetConfig(chain)
	bda, err := btcutil.DecodeAddress(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("btcutil.DecodeAddress(%s, %d) => %v", addr, chain, err)
	}
	script, err := txscript.PayToAddrScript(bda)
	if err != nil {
		return nil, fmt.Errorf("txscript.PayToAddrScript(%s, %d) => %v", addr, chain, err)
	}
	return script, nil
}

func ParseSequence(lock time.Duration, chain byte) (int64, error) {
	if lock < TimeLockMinimum || lock > TimeLockMaximum {
		return 0, fmt.Errorf("invalid lock %d", lock)
	}
	blockDuration := 10 * time.Minute
	switch chain {
	case ChainBitcoin:
	case ChainLitecoin:
		blockDuration = 150 * time.Second
	default:
	}
	// FIXME check litecoin timelock consensus as this may exceed 0xffff
	lock = lock / blockDuration
	if lock >= 0xffff {
		lock = 0xffff
	}
	return int64(lock), nil
}

func CheckFinalization(num uint64, coinbase bool) bool {
	if num >= uint64(chaincfg.MainNetParams.CoinbaseMaturity) {
		return true
	}
	return !coinbase && num >= TransactionConfirmations
}

func CheckDerivation(public string, chainCode []byte, maxRange uint32) error {
	for i := uint32(0); i <= maxRange; i++ {
		children := []uint32{i, i, i}
		_, _, err := DeriveBIP32(public, chainCode, children...)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeriveBIP32(public string, chainCode []byte, children ...uint32) (string, string, error) {
	key, err := hex.DecodeString(public)
	if err != nil {
		return "", "", err
	}
	parentFP := []byte{0x00, 0x00, 0x00, 0x00}
	version := []byte{0x04, 0x88, 0xb2, 0x1e}
	extPub := hdkeychain.NewExtendedKey(version, key, chainCode, parentFP, 0, 0, false)
	for _, i := range children {
		extPub, err = extPub.Derive(i)
		if err != nil {
			return "", "", err
		}
		if bytes.Equal(extPub.ChainCode(), chainCode) {
			return "", "", fmt.Errorf("invalid cc %x:%x", extPub.ChainCode(), chainCode)
		}
	}
	pub, err := extPub.ECPubKey()
	if err != nil {
		return "", "", err
	}
	return extPub.String(), hex.EncodeToString(pub.SerializeCompressed()), nil
}

func HashMessageForSignature(msg string, chain byte) ([]byte, error) {
	var buf bytes.Buffer
	prefix := "Bitcoin Signed Message:\n"
	switch chain {
	case ChainBitcoin:
	case ChainLitecoin:
		prefix = "Litecoin Signed Message:\n"
	default:
		return nil, fmt.Errorf("invalid chain %d", chain)
	}
	_ = wire.WriteVarString(&buf, 0, prefix)
	_ = wire.WriteVarString(&buf, 0, msg)
	return chainhash.DoubleHashB(buf.Bytes()), nil
}

func IsInsufficientInputError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "insufficient ")
}

func WriteBytes(enc *common.Encoder, b []byte) {
	enc.WriteInt(len(b))
	enc.Write(b)
}

func VerifyAddress(address string, coin uint32) error {
	_, err := btcutil.DecodeAddress(address, netParams(coin))
	return err
}

const (
	CoinBitcoin  = 0
	CoinLitecoin = 2
)

func netParams(coin uint32) *chaincfg.Params {
	switch coin {
	case CoinBitcoin:
		return &chaincfg.MainNetParams
	case CoinLitecoin:
		return &chaincfg.Params{
			Net:             0xdbb6c0fb,
			Bech32HRPSegwit: "ltc",

			PubKeyHashAddrID:        0x30,
			ScriptHashAddrID:        0x32,
			WitnessPubKeyHashAddrID: 0x06,
			WitnessScriptHashAddrID: 0x0A,

			HDPublicKeyID:  [4]byte{0x01, 0x9d, 0xa4, 0x64},
			HDPrivateKeyID: [4]byte{0x01, 0x9d, 0x9c, 0xfe},
		}
	default:
		panic(coin)
	}
}
