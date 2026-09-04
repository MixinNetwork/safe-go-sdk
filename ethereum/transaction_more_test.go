package ethereum

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	contractabi "github.com/MixinNetwork/go-safe-sdk/ethereum/abi"
	ga "github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNativeAndERC20Transactions(t *testing.T) {
	ctx := context.Background()
	native, err := CreateTransaction(
		ctx, TypeETHTx, 1, "operation-id", testSafeAddress,
		testDestination, "", "12345", big.NewInt(7),
	)
	require.NoError(t, err)
	assert.Equal(t, ethcommon.HexToAddress(testDestination), native.Destination)
	assert.Equal(t, big.NewInt(12345), native.Value)
	assert.Empty(t, native.Data)
	assert.Equal(t, uint8(operationTypeCall), native.Operation)
	assert.Len(t, native.Message, 32)
	assert.Len(t, native.TxHash, 64)
	assert.Equal(t, native.GetTransactionHash(), native.Message)
	assert.Equal(t, native.Hash("operation-id"), native.TxHash)

	outputs := native.ExtractOutputs()
	require.Len(t, outputs, 1)
	assert.Empty(t, outputs[0].TokenAddress)
	assert.Equal(t, testDestination, outputs[0].Destination)
	assert.Equal(t, big.NewInt(12345), outputs[0].Amount)

	fromOutputs, err := CreateTransactionFromOutputs(ctx, TypeETHTx, 1, "operation-id", testSafeAddress, []*Output{{
		Destination: testDestination,
		Amount:      big.NewInt(12345),
	}}, big.NewInt(7))
	require.NoError(t, err)
	assert.Equal(t, native.Message, fromOutputs.Message)

	erc20, err := CreateTransaction(
		ctx, TypeERC20Tx, 137, "operation-id", testSafeAddress,
		testDestination, testToken, "987654321", big.NewInt(8),
	)
	require.NoError(t, err)
	assert.Equal(t, ethcommon.HexToAddress(testToken), erc20.Destination)
	assert.Zero(t, erc20.Value.Sign())
	require.Len(t, erc20.Data, 68)
	assert.Equal(t, "a9059cbb", hex.EncodeToString(erc20.Data[:4]))

	outputs = erc20.ExtractOutputs()
	require.Len(t, outputs, 1)
	assert.Equal(t, testToken, outputs[0].TokenAddress)
	assert.Equal(t, testDestination, outputs[0].Destination)
	assert.Equal(t, big.NewInt(987654321), outputs[0].Amount)
}

func TestCreateTransactionErrors(t *testing.T) {
	ctx := context.Background()
	_, err := CreateTransaction(ctx, TypeETHTx, 1, "id", testSafeAddress, testDestination, "", "1", nil)
	assert.EqualError(t, err, "Invalid ethereum transaction nonce")
	_, err = CreateTransaction(ctx, TypeETHTx, 1, "id", testSafeAddress, testDestination, "", "not-an-integer", big.NewInt(1))
	assert.EqualError(t, err, "Fail to parse value to big.Int")
	_, err = CreateTransaction(ctx, TypeERC20Tx, 1, "id", testSafeAddress, testDestination, "invalid", "1", big.NewInt(1))
	assert.ErrorContains(t, err, "invalid ERC20 address")
	_, err = CreateTransaction(ctx, 99, 1, "id", testSafeAddress, testDestination, "", "1", big.NewInt(1))
	assert.EqualError(t, err, "invalid safe transaction type: 99")

	_, err = CreateTransactionFromOutputs(ctx, TypeETHTx, 1, "id", testSafeAddress, nil, big.NewInt(1))
	assert.EqualError(t, err, "invalid outputs to create safe transaction")
	_, err = CreateTransactionFromOutputs(ctx, TypeETHTx, 1, "id", testSafeAddress, []*Output{
		{Destination: testDestination, Amount: big.NewInt(1)},
		{Destination: testSafeAddress, Amount: big.NewInt(2)},
	}, big.NewInt(1))
	assert.EqualError(t, err, "invalid outputs to create safe transaction")
	_, err = CreateMultiSendTransaction(ctx, 1, "id", testSafeAddress, nil, nil)
	assert.EqualError(t, err, "Invalid ethereum transaction nonce")
	_, err = CreateEnableGuardTransaction(ctx, 1, "id", testSafeAddress, testDestination, nil)
	assert.ErrorContains(t, err, "invalid timelock")
}

func TestMultiSendTransactionRoundTrip(t *testing.T) {
	outputs := []*Output{
		{
			TokenAddress: EthereumEmptyAddress,
			Destination:  testDestination,
			Amount:       big.NewInt(100_000),
		},
		{
			TokenAddress: testToken,
			Destination:  testSafeAddress,
			Amount:       big.NewInt(200),
		},
	}
	tx, err := CreateTransactionFromOutputs(
		context.Background(), TypeMultiSendTx, 137, "operation-id",
		testSafeAddress, outputs, big.NewInt(9),
	)
	require.NoError(t, err)
	assert.Equal(t, uint8(operationTypeDelegateCall), tx.Operation)
	assert.Equal(t, ethcommon.HexToAddress(EthereumMultiSendAddress), tx.Destination)
	assert.Equal(t, GetMultiSendData(outputs), tx.Data)

	parsed, err := tx.ParseMultiSendData()
	require.NoError(t, err)
	require.Len(t, parsed, 2)
	assert.Equal(t, EthereumEmptyAddress, parsed[0].TokenAddress)
	assert.Equal(t, testDestination, parsed[0].Destination)
	assert.Equal(t, big.NewInt(100_000), parsed[0].Amount)
	assert.Equal(t, testToken, parsed[1].TokenAddress)
	assert.Equal(t, testSafeAddress, parsed[1].Destination)
	assert.Equal(t, big.NewInt(200), parsed[1].Amount)
	assert.Equal(t, parsed, tx.ExtractOutputs())

	raw := tx.Marshal()
	roundTrip, err := UnmarshalSafeTransaction(raw)
	require.NoError(t, err)
	assert.Equal(t, tx.TxHash, roundTrip.TxHash)
	assert.Equal(t, tx.ChainID, roundTrip.ChainID)
	assert.Equal(t, tx.SafeAddress, roundTrip.SafeAddress)
	assert.Equal(t, tx.Destination, roundTrip.Destination)
	assert.Equal(t, tx.Value, roundTrip.Value)
	assert.Equal(t, tx.Data, roundTrip.Data)
	assert.Equal(t, tx.Nonce, roundTrip.Nonce)
	assert.Equal(t, tx.Message, roundTrip.Message)
	assert.Equal(t, tx.Signatures, roundTrip.Signatures)

	_, err = (&SafeTransaction{Operation: operationTypeCall}).ParseMultiSendData()
	assert.ErrorContains(t, err, "invalid tx operation")
	_, err = (&SafeTransaction{Operation: operationTypeDelegateCall, Data: []byte{0, 0, 0, 0}}).ParseMultiSendData()
	assert.Error(t, err)
	_, err = (&SafeTransaction{Operation: operationTypeDelegateCall, Data: []byte{1, 2, 3}}).ParseMultiSendData()
	assert.ErrorContains(t, err, "invalid multi-send data length")

	multiSendABI, err := ga.JSON(strings.NewReader(contractabi.MultiSendMetaData.ABI))
	require.NoError(t, err)
	for _, malformed := range [][]byte{
		{1},
		append(make([]byte, 84), 1),
	} {
		data, packErr := multiSendABI.Pack("multiSend", malformed)
		require.NoError(t, packErr)
		_, err = (&SafeTransaction{Operation: operationTypeDelegateCall, Data: data}).ParseMultiSendData()
		assert.Error(t, err)
	}
}

func TestEnableGuardTransaction(t *testing.T) {
	tx, err := CreateEnableGuardTransaction(
		context.Background(), 1, "operation-id", testSafeAddress,
		testDestination, big.NewInt(3600),
	)
	require.NoError(t, err)
	assert.Equal(t, uint8(operationTypeDelegateCall), tx.Operation)
	assert.Equal(t, ethcommon.HexToAddress(EthereumMultiSendAddress), tx.Destination)
	assert.NotEmpty(t, tx.Data)

	outputs, err := tx.ParseMultiSendData()
	require.NoError(t, err)
	require.Len(t, outputs, 2)
	assert.Equal(t, testSafeAddress, outputs[0].Destination)
	assert.Equal(t, ethcommon.HexToAddress(EthereumSafeGuardAddress).Hex(), outputs[1].Destination)
}

func TestTransactionDataEncoding(t *testing.T) {
	amount := big.NewInt(123)
	erc20 := GetERC20TxData(testDestination, amount)
	require.Len(t, erc20, 68)
	assert.Equal(t, "a9059cbb", hex.EncodeToString(erc20[:4]))
	assert.Equal(t, ethcommon.HexToAddress(testDestination).Bytes(), erc20[16:36])
	assert.Equal(t, amount, new(big.Int).SetBytes(erc20[36:]))

	meta := GetMetaTxData(ethcommon.HexToAddress(testDestination), amount, erc20)
	require.Len(t, meta, 85+len(erc20))
	assert.Equal(t, byte(operationTypeCall), meta[0])
	assert.Equal(t, ethcommon.HexToAddress(testDestination).Bytes(), meta[1:21])
	assert.Equal(t, amount, new(big.Int).SetBytes(meta[21:53]))
	assert.Equal(t, int64(len(erc20)), new(big.Int).SetBytes(meta[53:85]).Int64())
	assert.Equal(t, erc20, meta[85:])
}

func TestMarshalAndUnmarshalErrors(t *testing.T) {
	_, err := UnmarshalSafeTransaction(nil)
	assert.Error(t, err)

	tx, err := CreateTransaction(
		context.Background(), TypeETHTx, 1, "id", testSafeAddress,
		testDestination, "", "1", big.NewInt(1),
	)
	require.NoError(t, err)
	raw := tx.Marshal()
	raw[len(raw)-1] = 'z'
	_, err = UnmarshalSafeTransaction(raw)
	assert.Error(t, err)

	tx.Signatures = [][]byte{{1}, {2}, {3}, {4}}
	_, err = UnmarshalSafeTransaction(tx.Marshal())
	assert.EqualError(t, err, "invalid signature count 4")

	assert.PanicsWithValue(t, "invalid safe transaction data", func() {
		(&SafeTransaction{Operation: operationTypeCall, Data: []byte{0, 0, 0, 0}}).ExtractOutputs()
	})
}

func TestSignAndVerifyTransaction(t *testing.T) {
	tx, err := CreateTransaction(
		context.Background(), TypeETHTx, 1, "id", testSafeAddress,
		testDestination, "", "1", big.NewInt(1),
	)
	require.NoError(t, err)
	raw := tx.Marshal()
	rawHex := hex.EncodeToString(raw)

	signatureHex, err := SignTx(rawHex, testPrivateKey)
	require.NoError(t, err)
	signature, err := hex.DecodeString(signatureHex)
	require.NoError(t, err)
	require.Len(t, signature, 65)
	assert.Contains(t, []byte{31, 32}, signature[64])

	private, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)
	public := hex.EncodeToString(crypto.CompressPubkey(&private.PublicKey))
	valid, err := CheckTransactionSignature(rawHex, public, signature)
	require.NoError(t, err)
	assert.True(t, valid)

	base64Signature, err := SignTx(base64.RawURLEncoding.EncodeToString(raw), testPrivateKey)
	require.NoError(t, err)
	assert.Equal(t, signatureHex, base64Signature)

	tx.Signatures[1] = signature
	signedRaw := hex.EncodeToString(tx.Marshal())
	assert.True(t, CheckTransactionPartiallySignedBy(signedRaw, public))
	assert.False(t, CheckTransactionPartiallySignedBy("invalid", public))
	assert.False(t, CheckTransactionPartiallySignedBy("00", public))
	otherPrivate, err := crypto.HexToECDSA("6cbed15c177e12c9e4974621d73cdd6e8f5e6473b6c7bfc94c9b5a8c4f6f6f6f")
	require.NoError(t, err)
	otherPublic := hex.EncodeToString(crypto.CompressPubkey(&otherPrivate.PublicKey))
	assert.False(t, CheckTransactionPartiallySignedBy(signedRaw, otherPublic))

	_, err = SignTx("not encoded", testPrivateKey)
	assert.Error(t, err)
	_, err = SignTx(rawHex, "invalid")
	assert.Error(t, err)
	_, err = CheckTransactionSignature("invalid", public, signature)
	assert.Error(t, err)

	processed := ProcessSignature(append(bytes.Repeat([]byte{0}, 64), 1))
	assert.Equal(t, byte(32), processed[64])
}
