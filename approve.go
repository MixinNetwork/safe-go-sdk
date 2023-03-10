package safe

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func ApproveSafeAccount(address, priv string) (string, error) {
	var buf bytes.Buffer
	_ = wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	_ = wire.WriteVarString(&buf, 0, address)
	hash := chainhash.DoubleHashB(buf.Bytes())
	b, err := hex.DecodeString(priv)
	if err != nil {
		return "", err
	}
	private, _ := btcec.PrivKeyFromBytes(b)
	sig := ecdsa.Sign(private, hash)
	return base64.RawURLEncoding.EncodeToString(sig.Serialize()), nil
}
