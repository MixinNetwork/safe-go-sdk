package operation

import (
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/bitcoin"
	"github.com/MixinNetwork/go-safe-sdk/ethereum"
)

func SignSafeTx(rawStr, privateStr string, chain byte) (string, error) {
	switch chain {
	case SafeChainBitcoin, SafeChainLitecoin:
		return bitcoin.SignTx(rawStr, privateStr, chain)
	case SafeChainEthereum, SafeChainMVM, SafeChainPolygon:
		return ethereum.SignTx(rawStr, privateStr)
	default:
		return "", fmt.Errorf("invalid chain: %d", chain)
	}
}

func HashMessageForSignature(msg string, chain byte) ([]byte, error) {
	switch chain {
	case SafeChainBitcoin, SafeChainLitecoin:
		return bitcoin.HashMessageForSignature(msg, chain)
	case SafeChainEthereum, SafeChainMVM, SafeChainPolygon:
		return ethereum.HashMessageForSignature(msg)
	default:
		return nil, fmt.Errorf("invalid chain: %d", chain)
	}
}

func CheckTransactionPartiallySignedBy(raw, public string, chain byte) bool {
	switch chain {
	case SafeChainBitcoin, SafeChainLitecoin:
		return bitcoin.CheckTransactionPartiallySignedBy(raw, public)
	case SafeChainEthereum, SafeChainMVM, SafeChainPolygon:
		return ethereum.CheckTransactionPartiallySignedBy(raw, public)
	default:
		return false
	}
}
