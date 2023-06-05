package operation

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

func SignSafeMessage(msg, priv string, chain byte) (string, error) {
	hash := HashMessageForSignature(msg, chain)
	b, err := hex.DecodeString(priv)
	if err != nil {
		return "", err
	}
	private, _ := btcec.PrivKeyFromBytes(b)
	sig := ecdsa.Sign(private, hash)
	return base64.RawURLEncoding.EncodeToString(sig.Serialize()), nil
}
