package bitcoin

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/txscript/v2"
	"github.com/btcsuite/btcd/wire/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func multisigScript(t *testing.T) []byte {
	t.Helper()
	builder := txscript.NewScriptBuilder().AddOp(txscript.OP_2)
	for i := byte(1); i <= 3; i++ {
		_, public := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{i}, 32))
		builder.AddData(public.SerializeCompressed())
	}
	script, err := builder.AddOp(txscript.OP_3).AddOp(txscript.OP_CHECKMULTISIG).Script()
	require.NoError(t, err)
	require.Greater(t, len(script), 100)
	return script
}

func testInput(t *testing.T, hashDigit byte, satoshi int64) *Input {
	t.Helper()
	return &Input{
		TransactionHash: strings.Repeat(string(hashDigit), 64),
		Index:           0,
		Satoshi:         satoshi,
		Script:          multisigScript(t),
		Sequence:        10,
	}
}

func TestInputAndOutputConstruction(t *testing.T) {
	assertType, err := checkScriptType(bytes.Repeat([]byte{1}, 33))
	require.NoError(t, err)
	assert.Equal(t, InputTypeP2WPKHAccoutant, assertType)
	assertType, err = checkScriptType(bytes.Repeat([]byte{1}, 101))
	require.NoError(t, err)
	assert.Equal(t, InputTypeP2WSHMultisigHolderSigner, assertType)
	_, err = checkScriptType([]byte{1, 2, 3})
	assert.ErrorContains(t, err, "invalid script")

	input := testInput(t, '1', 100_000)
	tx := wire.NewMsgTx(2)
	addr, err := addInput(tx, input, ChainBitcoin)
	require.NoError(t, err)
	assert.NotEmpty(t, addr)
	require.Len(t, tx.TxIn, 1)
	assert.Equal(t, uint32(MaxTransactionSequence), tx.TxIn[0].Sequence)

	backup := testInput(t, '2', 100_000)
	backup.RouteBackup = true
	backup.Sequence = 42
	tx = wire.NewMsgTx(2)
	_, err = addInput(tx, backup, ChainBitcoin)
	require.NoError(t, err)
	assert.Equal(t, uint32(42), tx.TxIn[0].Sequence)

	backup.Sequence = 0
	_, err = addInput(wire.NewMsgTx(2), backup, ChainBitcoin)
	assert.EqualError(t, err, "invalid sequence 0")
	bad := testInput(t, '3', 1)
	bad.TransactionHash = "invalid"
	_, err = addInput(wire.NewMsgTx(2), bad, ChainBitcoin)
	assert.Error(t, err)
	bad = testInput(t, '3', 1)
	bad.Script = []byte{1}
	_, err = addInput(wire.NewMsgTx(2), bad, ChainBitcoin)
	assert.ErrorContains(t, err, "invalid script")
	_, err = addInput(wire.NewMsgTx(2), testInput(t, '3', 1), 99)
	assert.EqualError(t, err, "invalid chain 99")

	receiver, _ := witnessAddress(t, ChainBitcoin, 0x55)
	tx = wire.NewMsgTx(2)
	require.NoError(t, addOutput(tx, receiver, 12_345, ChainBitcoin))
	require.Len(t, tx.TxOut, 1)
	assert.Equal(t, int64(12_345), tx.TxOut[0].Value)
	assert.Error(t, addOutput(tx, "invalid", 1, ChainBitcoin))
	assert.EqualError(t, addOutput(tx, receiver, 1, 99), "invalid chain 99")
}

func TestEstimateTransactionFee(t *testing.T) {
	main := testInput(t, '1', 100_000)
	receiver, _ := witnessAddress(t, ChainBitcoin, 0x66)
	outputs := []*Output{{Address: receiver, Satoshi: 50_000}}

	needed, err := EstimateTransactionFee(
		[]*Input{main},
		[]*Input{testInput(t, '2', 10_000)},
		outputs,
		10,
		[]byte("request-id"),
		ChainBitcoin,
	)
	require.NoError(t, err)
	assert.Zero(t, needed)

	needed, err = EstimateTransactionFee([]*Input{testInput(t, '1', 100_000)}, nil, outputs, 10, nil, ChainBitcoin)
	assert.ErrorContains(t, err, "insufficient fee")
	assert.Positive(t, needed)

	_, err = EstimateTransactionFee(
		[]*Input{testInput(t, '1', 10)}, nil,
		[]*Output{{Address: receiver, Satoshi: 11}}, 1, nil, ChainBitcoin,
	)
	assert.ErrorContains(t, err, "insufficient main")

	differentScript := testInput(t, '2', 10)
	differentScript.Script[2] ^= 1
	_, err = EstimateTransactionFee(
		[]*Input{testInput(t, '1', 10), differentScript}, nil, nil, 1, nil, ChainBitcoin,
	)
	assert.ErrorContains(t, err, "input address")

	_, err = EstimateTransactionFee(
		[]*Input{testInput(t, '1', 10)}, nil,
		[]*Output{{Address: "invalid", Satoshi: 1}}, 1, nil, ChainBitcoin,
	)
	assert.ErrorContains(t, err, "addOutput")
}

func TestBuildSignAndMarshalPartiallySignedTransaction(t *testing.T) {
	main := testInput(t, '1', 100_000)
	receiver, _ := witnessAddress(t, ChainBitcoin, 0x77)
	pst, err := BuildPartiallySignedTransaction(
		[]*Input{main},
		[]*Output{{Address: receiver, Satoshi: 40_000}},
		[]byte("request-id"),
		ChainBitcoin,
	)
	require.NoError(t, err)
	require.Len(t, pst.UnsignedTx.TxIn, 1)
	require.Len(t, pst.UnsignedTx.TxOut, 3)
	assert.NotEmpty(t, pst.Hash())

	hash, err := pst.SigHash(0)
	require.NoError(t, err)
	assert.Len(t, hash, 32)
	raw, err := pst.Marshal()
	require.NoError(t, err)
	roundTrip, err := UnmarshalPartiallySignedTransaction(raw)
	require.NoError(t, err)
	assert.Equal(t, pst.Hash(), roundTrip.Hash())

	private, public := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{9}, 32))
	signedHex, err := SignTx(hex.EncodeToString(raw), hex.EncodeToString(private.Serialize()), ChainBitcoin)
	require.NoError(t, err)
	assert.True(t, CheckTransactionPartiallySignedBy(signedHex, hex.EncodeToString(public.SerializeCompressed())))

	signedBase64, err := SignTx(base64.RawURLEncoding.EncodeToString(raw), hex.EncodeToString(private.Serialize()), ChainBitcoin)
	require.NoError(t, err)
	assert.True(t, CheckTransactionPartiallySignedBy(signedBase64, hex.EncodeToString(public.SerializeCompressed())))
	assert.False(t, CheckTransactionPartiallySignedBy("not-a-transaction", hex.EncodeToString(public.SerializeCompressed())))
	assert.False(t, CheckTransactionPartiallySignedBy("00", hex.EncodeToString(public.SerializeCompressed())))

	_, err = SignTx("not encoded", "private", ChainBitcoin)
	assert.Error(t, err)
	_, err = SignTx(hex.EncodeToString(raw), "invalid-private-key", ChainBitcoin)
	assert.Error(t, err)
	_, err = UnmarshalPartiallySignedTransaction([]byte("invalid"))
	assert.Error(t, err)

	wired, err := MarshalWiredTransaction(pst.UnsignedTx, wire.BaseEncoding, ChainBitcoin)
	require.NoError(t, err)
	assert.NotEmpty(t, wired)
	_, err = MarshalWiredTransaction(pst.UnsignedTx, wire.BaseEncoding, 99)
	assert.ErrorContains(t, err, "protocolVersion")
}

func TestBuildPartiallySignedTransactionErrors(t *testing.T) {
	receiver, _ := witnessAddress(t, ChainBitcoin, 0x44)
	_, err := BuildPartiallySignedTransaction(
		[]*Input{testInput(t, '1', 10)},
		[]*Output{{Address: receiver, Satoshi: 11}}, nil, ChainBitcoin,
	)
	assert.ErrorContains(t, err, "insufficient main")

	_, err = BuildPartiallySignedTransaction(
		[]*Input{testInput(t, '1', 10)},
		[]*Output{{Address: "invalid", Satoshi: 1}}, nil, ChainBitcoin,
	)
	assert.ErrorContains(t, err, "addOutput")
	_, err = BuildPartiallySignedTransaction(nil, nil, nil, 99)
	assert.EqualError(t, err, "invalid chain 99")
}

func TestVerifySignatureDER(t *testing.T) {
	private, public := btcec.PrivKeyFromBytes(bytes.Repeat([]byte{7}, 32))
	message := bytes.Repeat([]byte{3}, 32)
	signature := ecdsa.Sign(private, message).Serialize()
	publicHex := hex.EncodeToString(public.SerializeCompressed())

	assert.NoError(t, VerifySignatureDER(publicHex, message, signature))
	assert.Error(t, VerifySignatureDER(publicHex, bytes.Repeat([]byte{4}, 32), signature))
	assert.Error(t, VerifySignatureDER("invalid", message, signature))
	assert.Error(t, VerifySignatureDER(publicHex, message, []byte("invalid")))
}
