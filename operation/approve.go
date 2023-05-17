package operation

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

func ApproveSafeAccount(address, priv string, chain byte) (string, error) {
	hash := HashMessageForSignature(address, chain)
	b, err := hex.DecodeString(priv)
	if err != nil {
		return "", err
	}
	private, _ := btcec.PrivKeyFromBytes(b)
	sig := ecdsa.Sign(private, hash)
	return base64.RawURLEncoding.EncodeToString(sig.Serialize()), nil
}
