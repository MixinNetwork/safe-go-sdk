package operation

import (
	"github.com/MixinNetwork/go-safe-sdk/types"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const (
	BitcoinAssetId = "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
	PusdAssetId    = "31d2ea9c-95eb-3355-b65b-ba096853bc18"
)

func ProposeAccount(operationId, publicKey string, owners []string, threshold byte) *types.Operation {
	op := types.Operation{
		Id:     operationId,
		Type:   110,
		Curve:  1,
		Public: publicKey,
	}

	total := byte(len(owners))
	extra := []byte{threshold, total}
	for _, o := range owners {
		uid, err := uuid.FromString(o)
		if err != nil {
			panic(err)
		}
		extra = append(extra, uid.Bytes()...)
	}
	op.Extra = extra
	return &op
}

func ProposeTransaction(operationId, publicKey string, head, destination string) *types.Operation {
	extra := uuid.FromStringOrNil(head).Bytes()
	extra = append(extra, []byte(destination)...)
	op := &types.Operation{
		Id:     operationId,
		Type:   112,
		Curve:  1,
		Public: publicKey,
		Extra:  extra,
	}
	return op
}

func BuildTransfer(assetId, amount, operationId, memo string) (*mixin.TransferInput, error) {
	a, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, err
	}
	input := &mixin.TransferInput{
		AssetID: assetId,
		Amount:  a,
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
	return input, nil
}
