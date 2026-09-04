package ethereum

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil/v2/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPrivateKey  = "04f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f2b28f6dcb7ea9caf"
	testSafeAddress = "0x346607eb15821A4E194628444F3705c26C8E6eBe"
	testDestination = "0xA03A8590BB3A2cA5c747c8b99C63DA399424a055"
	testToken       = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
)

func TestAssetKeysAndIDs(t *testing.T) {
	valid := "0xc2132d05d31c914a87c6611c10748aeb04b58e8f"
	assert.NoError(t, VerifyAssetKey(valid))
	for _, invalid := range []string{
		"0x1234",
		"c2132d05d31c914a87c6611c10748aeb04b58e8f",
		"0xC2132d05d31c914a87c6611c10748aeb04b58e8f",
		"0xzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
	} {
		t.Run(invalid, func(t *testing.T) {
			assert.Error(t, VerifyAssetKey(invalid))
		})
	}

	id := GenerateAssetId(ChainPolygon, strings.ToUpper(valid))
	parsed := uuid.Must(uuid.FromString(id))
	assert.Equal(t, byte(3), parsed.Version())
	assert.Equal(t, id, BuildChainAssetId(GetMixinChainID(ChainPolygon), valid))
	assert.Panics(t, func() { GenerateAssetId(99, valid) })
	assert.Panics(t, func() { GenerateAssetId(ChainEthereum, "invalid") })
}

func TestAmountsAndChainIDs(t *testing.T) {
	assert.Equal(t, big.NewInt(1_250_000), ParseAmount("1.25", 6))
	assert.Equal(t, "1.25", UnitAmount(big.NewInt(1_250_000), 6))
	assert.Equal(t, "-0.5", UnitAmount(big.NewInt(-5), 1))
	assert.Panics(t, func() { ParseAmount("not-a-number", 18) })
	assert.Panics(t, func() { ParseAmount("0.0000001", 6) })

	assert.Equal(t, int64(1), GetEvmChainID(ChainEthereum))
	assert.Equal(t, int64(73927), GetEvmChainID(ChainMVM))
	assert.Equal(t, int64(137), GetEvmChainID(ChainPolygon))
	assert.Equal(t, "43d61dcd-e413-450d-80b8-101d5e903357", GetMixinChainID(ChainEthereum))
	assert.Equal(t, "a0ffd769-5850-4b48-9651-d2ae44a3e64d", GetMixinChainID(ChainMVM))
	assert.Equal(t, "b7938396-3f94-4e0a-9179-d3440718156f", GetMixinChainID(ChainPolygon))
	assert.PanicsWithValue(t, int64(99), func() { GetEvmChainID(99) })
	assert.PanicsWithValue(t, int64(99), func() { GetMixinChainID(99) })
}

func TestEthereumAddressesAndKeys(t *testing.T) {
	assert.Equal(t, testDestination, NormalizeAddress(testDestination))
	assert.Equal(t, testDestination, NormalizeAddress("0xa03a8590bb3a2ca5c747c8b99c63da399424a055"))
	assert.Empty(t, NormalizeAddress(EthereumEmptyAddress))
	assert.Empty(t, NormalizeAddress("not-an-address"))

	private, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)
	wantAddress := crypto.PubkeyToAddress(private.PublicKey)
	gotAddress, err := PrivToAddress(testPrivateKey)
	require.NoError(t, err)
	assert.Equal(t, wantAddress, *gotAddress)
	_, err = PrivToAddress("invalid")
	assert.Error(t, err)

	compressed := hex.EncodeToString(crypto.CompressPubkey(&private.PublicKey))
	parsedAddress, err := ParseEthereumCompressedPublicKey(compressed)
	require.NoError(t, err)
	assert.Equal(t, wantAddress, *parsedAddress)
	assert.NoError(t, VerifyHolderKey(compressed))
	assert.Error(t, VerifyHolderKey("invalid"))

	master, err := hdkeychain.NewMaster(bytes.Repeat([]byte{8}, 32), &chaincfg.MainNetParams)
	require.NoError(t, err)
	publicMaster, err := master.Neuter()
	require.NoError(t, err)
	extendedAddress, err := ParseEthereumUncompressedPublicKey(publicMaster.String())
	require.NoError(t, err)
	ecPublic, err := publicMaster.ECPubKey()
	require.NoError(t, err)
	decompressed, err := crypto.DecompressPubkey(ecPublic.SerializeCompressed())
	require.NoError(t, err)
	assert.Equal(t, crypto.PubkeyToAddress(*decompressed), *extendedAddress)
}

func TestEthereumMessageSignatures(t *testing.T) {
	message := []byte("safe message")
	hash, err := HashMessageForSignature(hex.EncodeToString(message))
	require.NoError(t, err)
	assert.Len(t, hash, 32)
	_, err = HashMessageForSignature("invalid hex")
	assert.Error(t, err)

	private, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)
	signature, err := crypto.Sign(hash, private)
	require.NoError(t, err)
	public := hex.EncodeToString(crypto.CompressPubkey(&private.PublicKey))
	assert.NoError(t, VerifyHashSignature(public, hash, signature))
	assert.NoError(t, VerifyMessageSignature(public, message, signature))
	assert.Error(t, VerifyHashSignature(public, bytes.Repeat([]byte{1}, 32), signature))
	assert.Error(t, VerifyHashSignature("not-hex", hash, signature))
	assert.EqualError(t, VerifyHashSignature(public, hash, []byte{1}), "invalid signature length 1")
	_, err = ParseEthereumCompressedPublicKey("zz")
	assert.Error(t, err)
	_, err = ParseEthereumUncompressedPublicKey("invalid")
	assert.Error(t, err)

	bytes32 := toBytes32(bytes.Repeat([]byte{0xab}, 32))
	assert.Equal(t, bytes.Repeat([]byte{0xab}, 32), bytes32[:])
	address, err := ParseEthereumCompressedPublicKey(public)
	require.NoError(t, err)
	assert.Equal(t, crypto.PubkeyToAddress(private.PublicKey), *address)
}
