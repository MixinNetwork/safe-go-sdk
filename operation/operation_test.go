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
	hash, err := HashMessageForSignature("hello crypto", 1)
	assert.Nil(err)
	err = VerifySafeMessage(hex.EncodeToString(pub.SerializeCompressed()), hash, sigBuf, 1)
	assert.Nil(err)

	// APPROVE:ACCOUNT-ID:ADDRESS
	sigBuf, _ = base64.RawURLEncoding.DecodeString("MEUCIQDpVx9JXZDNTp17E1LVBbD7BSkrNxw4V5Y00z0jsY2oDQIgKGk7RVIjO2NEvoLn5fotX7j4Kc4YdYI3ueq7GlmAWyc")
	hash, err = HashMessageForSignature("APPROVE:8aef8130-aa9c-418a-871d-e920fed2f0e4:bc1qd4qjpy8n3gksd56aqc9pz36tpy26ev2sq93xwkz3qjk64dra8ruq5p5hyv", 1)
	assert.Nil(err)
	err = VerifySafeMessage("0339af9aed5542535f1c609d45847ddc56d0d469cb59a3bcddf6555e028e42457e", hash, sigBuf, 1)
	assert.Nil(err)

	// APPROVE|REVOKE:TX-ID:TX-HASH
	sigBuf, _ = base64.RawURLEncoding.DecodeString("MEQCIE5JQAc8yY6RtN1WXl4FpSSKT66ck1Vs397g-BoGPHVAAiByaZ5hviSmsN1wiWPUhyetsC4wqpPiYFOlplWfylzRVA")
	hash, err = HashMessageForSignature("APPROVE:3ec57759-4bc9-4084-99f2-c712f1da31db:eaac19b6879b99cfab35b6f7f52d421eebaf2256f7ec4b3cdde2c1240bbe63ff", 1)
	assert.Nil(err)
	err = VerifySafeMessage("02b06814acd1b5993450c3732e4d8d5be8a19bf102461a5ccdfcc4926bed504b0a", hash, sigBuf, 1)
	assert.Nil(err)

	holder := "03911c1ef3960be7304596cfa6073b1d65ad43b421a4c272142cc7a8369b510c56"
	receiver := "bc1ql0up0wwazxt6xlj84u9fnvhnagjjetcn7h4z5xxvd0kf5xuczjgqq2aehc"
	op, err := ProposeInheritanceTransaction("358c0e9e-8d9c-4e0f-acde-8945a859763a", holder, TransactionTypeSetInheritance, "ce8491f2-3fde-4d2e-a4cc-4fcf707889c3", receiver, 1, "", "6e85e33c27105143808a5fdcdea96f3ef7cb2d8553fcb7680b10c3778c55059a", 60000)
	assert.Nil(err)
	assert.Equal(
		"036e85e33c27105143808a5fdcdea96f3ef7cb2d8553fcb7680b10c3778c55059aea60ce8491f23fde4d2ea4cc4fcf707889c3626331716c307570307777617a787436786c6a38347539666e76686e61676a6a6574636e3768347a3578787664306b66357875637a6a6771713261656863",
		hex.EncodeToString(op.Extra),
	)
	assert.Equal(
		"358c0e9e8d9c4e0facde8945a859763a70012103911c1ef3960be7304596cfa6073b1d65ad43b421a4c272142cc7a8369b510c5671036e85e33c27105143808a5fdcdea96f3ef7cb2d8553fcb7680b10c3778c55059aea60ce8491f23fde4d2ea4cc4fcf707889c3626331716c307570307777617a787436786c6a38347539666e76686e61676a6a6574636e3768347a3578787664306b66357875637a6a6771713261656863",
		hex.EncodeToString(op.Encode()),
	)

	op, err = ProposeInheritanceTransaction("1924a324-dbcb-48db-b0ea-5d23ebe59471", holder, TransactionTypeRemoveInheritance, "a229d702-1888-46b6-9141-f875ffe6c566", receiver, 1, "af36f755-a48a-3408-8a97-092007f9e2d2", "8f59870d883f2fa420dbcc2ce6fdfac72076b63982bdf8fe31aaa3b642845a7f", 20)
	assert.Nil(err)
	assert.Equal(
		"04af36f755a48a34088a97092007f9e2d2a229d702188846b69141f875ffe6c566626331716c307570307777617a787436786c6a38347539666e76686e61676a6a6574636e3768347a3578787664306b66357875637a6a6771713261656863",
		hex.EncodeToString(op.Extra),
	)
	assert.Equal(
		"1924a324dbcb48dbb0ea5d23ebe5947170012103911c1ef3960be7304596cfa6073b1d65ad43b421a4c272142cc7a8369b510c565f04af36f755a48a34088a97092007f9e2d2a229d702188846b69141f875ffe6c566626331716c307570307777617a787436786c6a38347539666e76686e61676a6a6574636e3768347a3578787664306b66357875637a6a6771713261656863",
		hex.EncodeToString(op.Encode()),
	)
}
