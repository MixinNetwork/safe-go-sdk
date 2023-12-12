package operation

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/bitcoin"
	"github.com/MixinNetwork/go-safe-sdk/ethereum"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
)

func SignSafeMessage(msg, priv string, chain byte) (string, error) {
	switch chain {
	case SafeChainBitcoin, SafeChainLitecoin:
		hash, err := HashMessageForSignature(msg, chain)
		if err != nil {
			return "", err
		}
		b, err := hex.DecodeString(priv)
		if err != nil {
			return "", err
		}
		private, _ := btcec.PrivKeyFromBytes(b)
		sig := ecdsa.Sign(private, hash)
		return base64.RawURLEncoding.EncodeToString(sig.Serialize()), nil
	case SafeChainEthereum, SafeChainMVM:
		hash, err := HashMessageForSignature(msg, chain)
		if err != nil {
			return "", err
		}
		private, err := crypto.HexToECDSA(priv)
		if err != nil {
			return "", err
		}
		sig, err := crypto.Sign(hash, private)
		if err != nil {
			return "", err
		}
		sig = ethereum.ProcessSignature(sig)
		return base64.RawURLEncoding.EncodeToString(sig), nil
	default:
		return "", fmt.Errorf("invalid chain: %d", chain)
	}
}

func VerifySafeMessage(public string, msg, sig []byte, chain byte) error {
	switch chain {
	case SafeChainBitcoin, SafeChainLitecoin:
		return bitcoin.VerifySignatureDER(public, msg, sig)
	case SafeChainEthereum, SafeChainMVM:
		return ethereum.VerifyMessageSignature(public, msg, sig)
	default:
		return fmt.Errorf("invalid chain: %d", chain)
	}
}
