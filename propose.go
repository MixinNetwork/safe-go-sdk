package safe

import (
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const (
	BitcoinAssetId = "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
)

func ProposeAccount(operationId, publicKey string, owners []string) (string, string, error) {
	op := &Operation{
		Id:     operationId,
		Type:   110,
		Curve:  1,
		Public: publicKey,
	}

	threshold := byte(1)
	total := byte(1)
	extra := []byte{threshold, total}
	// TODO support multi
	uid := uuid.FromStringOrNil(owners[0])
	op.Extra = append(extra, uid.Bytes()...)

	memo := base64.RawURLEncoding.EncodeToString(op.Encode())
	return op.Id, memo, nil
}

func BuildTransfer(operationId, memo string) *mixin.TransferInput {
	input := &mixin.TransferInput{
		AssetID: BitcoinAssetId,
		Amount:  decimal.NewFromFloat(0.0001),
		TraceID: operationId,
		Memo:    memo,
	}
	input.OpponentMultisig.Receivers = []string{
		"71b72e67-3636-473a-9ee4-db7ba3094057",
		"148e696f-f1db-4472-a907-ceea50c5cfde",
		"c9a9a719-4679-4057-bcf0-98945ed95a81",
		"b45dcee0-23d7-4ad1-b51e-c681a257c13e",
		"fcb87491-4fa0-4c2f-b387-262b63cbc112",
	}
	input.OpponentMultisig.Threshold = 4
	return input
}
