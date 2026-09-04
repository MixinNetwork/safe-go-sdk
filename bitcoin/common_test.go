package bitcoin

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	safecommon "github.com/MixinNetwork/go-safe-sdk/common"
	mixincommon "github.com/MixinNetwork/mixin/common"
	"github.com/btcsuite/btcd/address/v2"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/v2/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/btcsuite/btcd/txscript/v2"
	"github.com/btcsuite/btcd/wire/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func witnessAddress(t *testing.T, chain byte, fill byte) (string, []byte) {
	t.Helper()
	params, err := safecommon.NetConfig(chain)
	require.NoError(t, err)
	addr, err := address.NewAddressWitnessPubKeyHash(bytes.Repeat([]byte{fill}, 20), params)
	require.NoError(t, err)
	script, err := txscript.PayToAddrScript(addr)
	require.NoError(t, err)
	return addr.EncodeAddress(), script
}

func TestChainParameters(t *testing.T) {
	assert.Equal(t, []byte{2, 0, 0, 0}, BitcoinDefaultDerivationPath())

	value, err := ValueDust(ChainBitcoin)
	require.NoError(t, err)
	assert.Equal(t, int64(1_000), value)
	value, err = ValueDust(ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, int64(10_000), value)
	_, err = ValueDust(99)
	assert.EqualError(t, err, "invalid chain 99")

	version, err := protocolVersion(ChainBitcoin)
	require.NoError(t, err)
	assert.Equal(t, uint32(wire.ProtocolVersion), version)
	version, err = protocolVersion(ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, uint32(70015), version)
	_, err = protocolVersion(99)
	assert.EqualError(t, err, "invalid chain 99")

	assert.Same(t, &chaincfg.MainNetParams, netParams(CoinBitcoin))
	assert.Equal(t, "ltc", netParams(CoinLitecoin).Bech32HRPSegwit)
	assert.PanicsWithValue(t, uint32(99), func() { netParams(99) })
}

func TestParseSatoshi(t *testing.T) {
	tests := []struct {
		amount string
		want   int64
	}{
		{amount: "0", want: 0},
		{amount: "1", want: ValueSatoshi},
		{amount: "0.00000001", want: 1},
		{amount: "-1.25", want: -125_000_000},
		{amount: "21000000.00000000", want: 2_100_000_000_000_000},
	}
	for _, tt := range tests {
		t.Run(tt.amount, func(t *testing.T) {
			got, err := ParseSatoshi(tt.amount)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	for _, amount := range []string{"nope", "0.000000001", "92233720369"} {
		t.Run("invalid_"+amount, func(t *testing.T) {
			_, err := ParseSatoshi(amount)
			assert.Error(t, err)
		})
	}
}

func TestAddresses(t *testing.T) {
	bitcoinAddress, bitcoinScript := witnessAddress(t, ChainBitcoin, 0x11)
	litecoinAddress, litecoinScript := witnessAddress(t, ChainLitecoin, 0x22)

	parsed, err := ParseAddress(bitcoinAddress, ChainBitcoin)
	require.NoError(t, err)
	assert.Equal(t, bitcoinScript, parsed)
	parsed, err = ParseAddress(litecoinAddress, ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, litecoinScript, parsed)

	assert.NoError(t, VerifyAddress(bitcoinAddress, CoinBitcoin))
	assert.NoError(t, VerifyAddress(litecoinAddress, CoinLitecoin))
	_, err = ParseAddress("not-an-address", ChainBitcoin)
	assert.ErrorContains(t, err, "bitcoin.VerifyAddress")
	_, err = ParseAddress(bitcoinAddress, 99)
	assert.EqualError(t, err, "ParseAddress("+bitcoinAddress+", 99)")

	extracted, err := ExtractPkScriptAddr(bitcoinScript, ChainBitcoin)
	require.NoError(t, err)
	assert.Equal(t, bitcoinAddress, extracted)
	extracted, err = ExtractPkScriptAddr(litecoinScript, ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, litecoinAddress, extracted)
	_, err = ExtractPkScriptAddr([]byte{txscript.OP_TRUE}, ChainBitcoin)
	assert.ErrorContains(t, err, "unsupported pkscript")
	_, err = ExtractPkScriptAddr(bitcoinScript, 99)
	assert.EqualError(t, err, "invalid chain 99")
}

func TestParseSequenceAndFinalization(t *testing.T) {
	sequence, err := ParseSequence(time.Hour, ChainBitcoin)
	require.NoError(t, err)
	assert.Equal(t, int64(6), sequence)
	sequence, err = ParseSequence(time.Hour, ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, int64(24), sequence)
	sequence, err = ParseSequence(TimeLockMaximum, ChainLitecoin)
	require.NoError(t, err)
	assert.Equal(t, int64(0xffff), sequence)

	_, err = ParseSequence(TimeLockMinimum-time.Second, ChainBitcoin)
	assert.Error(t, err)
	_, err = ParseSequence(TimeLockMaximum+time.Second, ChainBitcoin)
	assert.Error(t, err)

	assert.False(t, CheckFinalization(0, false))
	assert.True(t, CheckFinalization(TransactionConfirmations, false))
	assert.False(t, CheckFinalization(uint64(chaincfg.MainNetParams.CoinbaseMaturity-1), true))
	assert.True(t, CheckFinalization(uint64(chaincfg.MainNetParams.CoinbaseMaturity), true))
}

func TestBIP32Derivation(t *testing.T) {
	_, publicKey := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{0x42}, 32))
	public := hex.EncodeToString(publicKey.SerializeCompressed())
	chainCode := bytes.Repeat([]byte{0x24}, 32)

	xpub, derived, err := DeriveBIP32(public, chainCode, 0, 1, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, xpub)
	require.Len(t, derived, 66)
	_, err = btcec.ParsePubKey(safecommon.DecodeHexOrPanic(derived))
	assert.NoError(t, err)

	withPath, err := DeriveBIP32WithPath(public, hex.EncodeToString(chainCode), []byte{3, 0, 1, 2})
	require.NoError(t, err)
	assert.Equal(t, derived, withPath)
	assert.NoError(t, CheckDerivation(public, chainCode, 2))

	_, _, err = DeriveBIP32("not-hex", chainCode, 0)
	assert.Error(t, err)
	_, _, err = DeriveBIP32(public, chainCode, hdkeychain.HardenedKeyStart)
	assert.Error(t, err)
	_, err = DeriveBIP32WithPath(public, "not-hex", []byte{0})
	assert.Error(t, err)
	assert.PanicsWithValue(t, byte(4), func() {
		_, _ = DeriveBIP32WithPath(public, hex.EncodeToString(chainCode), []byte{4})
	})
}

func TestMessageAndEncodingHelpers(t *testing.T) {
	bitcoinHash, err := HashMessageForSignature("hello", ChainBitcoin)
	require.NoError(t, err)
	litecoinHash, err := HashMessageForSignature("hello", ChainLitecoin)
	require.NoError(t, err)
	assert.Len(t, bitcoinHash, 32)
	assert.Len(t, litecoinHash, 32)
	assert.NotEqual(t, bitcoinHash, litecoinHash)
	_, err = HashMessageForSignature("hello", 99)
	assert.EqualError(t, err, "invalid chain 99")

	assert.False(t, IsInsufficientInputError(nil))
	assert.False(t, IsInsufficientInputError(errors.New("other error")))
	assert.True(t, IsInsufficientInputError(errors.New("insufficient main 1 2")))

	enc := mixincommon.NewEncoder()
	WriteBytes(enc, []byte("payload"))
	dec := mixincommon.NewDecoder(enc.Bytes())
	length, err := dec.ReadInt()
	require.NoError(t, err)
	assert.Equal(t, len("payload"), length)
	buf := make([]byte, length)
	require.NoError(t, dec.Read(buf))
	assert.Equal(t, "payload", string(buf))
}
