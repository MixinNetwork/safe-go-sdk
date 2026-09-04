package types

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationRoundTrip(t *testing.T) {
	operation := &Operation{
		Id:     "67b91655-5d09-4686-a2e2-6419388e96e4",
		Type:   OperationTypeSignInput,
		Curve:  CurveSecp256k1ECDSAEthereum,
		Public: "02aabbcc",
		Extra:  []byte("transaction data"),
	}

	encoded := operation.Encode()
	decoded, err := DecodeOperation(encoded)
	require.NoError(t, err)
	assert.Equal(t, operation, decoded)
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(encoded), operation.EncodeBase64())
	assert.Equal(t, uuid.Must(uuid.FromString(operation.Id)).Bytes(), operation.IdBytes())

	const appID = "8aef8130-aa9c-418a-871d-e920fed2f0e4"
	memo, err := base64.RawURLEncoding.DecodeString(operation.EncodeMtgMemo(appID))
	require.NoError(t, err)
	assert.Equal(t, uuid.Must(uuid.FromString(appID)).Bytes(), memo[:16])
	assert.Equal(t, encoded, memo[16:])
}

func TestOperationWithEmptyFields(t *testing.T) {
	operation := &Operation{Id: "67b91655-5d09-4686-a2e2-6419388e96e4"}
	decoded, err := DecodeOperation(operation.Encode())
	require.NoError(t, err)
	assert.Empty(t, decoded.Public)
	assert.Nil(t, decoded.Extra)
}

func TestDecodeOperationErrors(t *testing.T) {
	valid := (&Operation{
		Id:     "67b91655-5d09-4686-a2e2-6419388e96e4",
		Type:   OperationTypeKeygenInput,
		Curve:  CurveSecp256k1ECDSABitcoin,
		Public: "0102",
		Extra:  []byte("extra"),
	}).Encode()

	for _, length := range []int{0, 15, 16, 17, 18, 19, 21, len(valid) - 1} {
		t.Run(string(rune(length)), func(t *testing.T) {
			_, err := DecodeOperation(valid[:length])
			assert.Error(t, err)
		})
	}
}

func TestOperationPanicsForInvalidInput(t *testing.T) {
	assert.Panics(t, func() {
		(&Operation{Id: "invalid"}).Encode()
	})
	assert.PanicsWithValue(t, "not-hex", func() {
		(&Operation{Id: "67b91655-5d09-4686-a2e2-6419388e96e4", Public: "not-hex"}).Encode()
	})
	assert.PanicsWithValue(t, 201, func() {
		(&Operation{
			Id:    "67b91655-5d09-4686-a2e2-6419388e96e4",
			Extra: bytes.Repeat([]byte{1}, 201),
		}).Encode()
	})
	assert.Panics(t, func() {
		(&Operation{Id: "invalid"}).IdBytes()
	})
	assert.PanicsWithValue(t, "not-hex", func() {
		DecodeHex("not-hex")
	})
}
