package common

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetConfig(t *testing.T) {
	bitcoin, err := NetConfig(ChainBitcoin)
	require.NoError(t, err)
	assert.Same(t, &chaincfg.MainNetParams, bitcoin)

	litecoin, err := NetConfig(ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, "ltc", litecoin.Bech32HRPSegwit)
	assert.Equal(t, byte(0x30), litecoin.PubKeyHashAddrID)
	assert.Equal(t, byte(0x32), litecoin.ScriptHashAddrID)
	assert.Equal(t, [4]byte{0x01, 0x9d, 0xa4, 0x64}, litecoin.HDPublicKeyID)

	params, err := NetConfig(255)
	assert.Nil(t, params)
	assert.EqualError(t, err, "invalid chain 255")
}

func TestDecodeHexOrPanic(t *testing.T) {
	assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, DecodeHexOrPanic("deadbeef"))
	assert.PanicsWithValue(t, "not-hex", func() {
		DecodeHexOrPanic("not-hex")
	})
}

func TestEncodeMixinExtra(t *testing.T) {
	const (
		appID = "8aef8130-aa9c-418a-871d-e920fed2f0e4"
		memo  = "safe memo"
	)

	encoded := EncodeMixinExtra(appID, memo)
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	require.NoError(t, err)
	require.Len(t, decoded, 16+len(memo))
	assert.Equal(t, uuid.Must(uuid.FromString(appID)).Bytes(), decoded[:16])
	assert.Equal(t, memo, string(decoded[16:]))

	assert.Panics(t, func() {
		EncodeMixinExtra("invalid", memo)
	})
}
