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

	// APPROVE:ACCOUNT-ID:ADDRESS
	sigBuf, _ = base64.RawURLEncoding.DecodeString("MEUCIQDpVx9JXZDNTp17E1LVBbD7BSkrNxw4V5Y00z0jsY2oDQIgKGk7RVIjO2NEvoLn5fotX7j4Kc4YdYI3ueq7GlmAWyc")
	hash = HashMessageForSignature("APPROVE:8aef8130-aa9c-418a-871d-e920fed2f0e4:bc1qd4qjpy8n3gksd56aqc9pz36tpy26ev2sq93xwkz3qjk64dra8ruq5p5hyv", 1)
	err = VerifySafeMessage("0339af9aed5542535f1c609d45847ddc56d0d469cb59a3bcddf6555e028e42457e", hash, sigBuf)

	// APPROVE|REVOKE:TX-ID:TX-HASH
	sigBuf, _ = base64.RawURLEncoding.DecodeString("MEQCIE5JQAc8yY6RtN1WXl4FpSSKT66ck1Vs397g-BoGPHVAAiByaZ5hviSmsN1wiWPUhyetsC4wqpPiYFOlplWfylzRVA")
	hash = HashMessageForSignature("APPROVE:3ec57759-4bc9-4084-99f2-c712f1da31db:eaac19b6879b99cfab35b6f7f52d421eebaf2256f7ec4b3cdde2c1240bbe63ff", 1)
	err = VerifySafeMessage("02b06814acd1b5993450c3732e4d8d5be8a19bf102461a5ccdfcc4926bed504b0a", hash, sigBuf)
}
