package operation

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
)

func TestOperation(t *testing.T) {
	assert := assert.New(t)

	seed, _ := hex.DecodeString("e66cdf44cb43927c3dd9288f5d3efb11f37fe68d791430c13a3a17492baa4724")
	priv, pub := btcec.PrivKeyFromBytes(seed)
	assert.Len(priv.Serialize(), 32)
	assert.Len(pub.SerializeCompressed(), 33)

	sig, err := SignSafeMessage("hello crypto", hex.EncodeToString(priv.Serialize()), 1)
	assert.Nil(err)
	assert.Equal("MEQCIDy5QeU_AjIMWZcZSA564scbrOipplGVjrSyh_xF-2qUAiAff7_Rb0MViZQe4sQ5_Aai0WMQiI40vqQ3RrU1FmlW9A", sig)

	sigBuf, _ := base64.RawURLEncoding.DecodeString(sig)
	hash := HashMessageForSignature("hello crypto", 1)
	err = VerifySafeMessage(hex.EncodeToString(pub.SerializeCompressed()), hash, sigBuf)
	assert.Nil(err)
}
