package operation

import (
	"encoding/binary"
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/types"
	"github.com/gofrs/uuid/v5"
)

const (
	BitcoinAssetId = "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
	PusdAssetId    = "31d2ea9c-95eb-3355-b65b-ba096853bc18"

	CurveSecp256k1ECDSABitcoin  = 1
	CurveSecp256k1ECDSAEthereum = 2
	CurveSecp256k1ECDSALitecoin = 100 + CurveSecp256k1ECDSABitcoin
	CurveSecp256k1ECDSAMVM      = 100 + CurveSecp256k1ECDSAEthereum

	// For all Bitcoin like chains
	ActionBitcoinSafeProposeAccount     = 110
	ActionBitcoinSafeApproveAccount     = 111
	ActionBitcoinSafeProposeTransaction = 112
	ActionBitcoinSafeApproveTransaction = 113
	ActionBitcoinSafeRevokeTransaction  = 114
	ActionBitcoinSafeCloseAccount       = 115

	// For all Ethereum like chains
	ActionEthereumSafeProposeAccount     = 130
	ActionEthereumSafeApproveAccount     = 131
	ActionEthereumSafeProposeTransaction = 132
	ActionEthereumSafeApproveTransaction = 133
	ActionEthereumSafeRevokeTransaction  = 134
	ActionEthereumSafeCloseAccount       = 135
	ActionEthereumSafeRefundTransaction  = 136

	TransactionTypeNormal   = 0
	TransactionTypeRecovery = 1
)

func ProposeAccount(operationId, publicKey string, owners []string, threshold, chain byte, timeLock uint16) (*types.Operation, error) {
	var action, curve uint8
	switch chain {
	case CurveSecp256k1ECDSABitcoin:
		action = ActionBitcoinSafeProposeAccount
		curve = CurveSecp256k1ECDSABitcoin
	case CurveSecp256k1ECDSALitecoin:
		action = ActionBitcoinSafeProposeAccount
		curve = CurveSecp256k1ECDSALitecoin
	case CurveSecp256k1ECDSAEthereum:
		action = ActionEthereumSafeProposeAccount
		curve = CurveSecp256k1ECDSAEthereum
	case CurveSecp256k1ECDSAMVM:
		action = ActionEthereumSafeProposeAccount
		curve = CurveSecp256k1ECDSAMVM
	default:
		return nil, fmt.Errorf("invalid chain: %d", chain)
	}

	op := types.Operation{
		Id:     operationId,
		Type:   action,
		Curve:  curve,
		Public: publicKey,
	}

	timelock := binary.BigEndian.AppendUint16(nil, timeLock)
	total := byte(len(owners))
	extra := append(timelock, threshold, total)
	for _, o := range owners {
		uid, err := uuid.FromString(o)
		if err != nil {
			return nil, fmt.Errorf("invalid uuid %s", o)
		}
		extra = append(extra, uid.Bytes()...)
	}
	op.Extra = extra
	return &op, nil
}

func ProposeTransaction(operationId, publicKey string, typ byte, head, destination string, chain byte, assetId string) (*types.Operation, error) {
	var action, curve uint8
	switch chain {
	case CurveSecp256k1ECDSABitcoin:
		action = ActionBitcoinSafeProposeTransaction
		curve = CurveSecp256k1ECDSABitcoin
	case CurveSecp256k1ECDSALitecoin:
		action = ActionBitcoinSafeProposeTransaction
		curve = CurveSecp256k1ECDSALitecoin
	case CurveSecp256k1ECDSAEthereum:
		action = ActionEthereumSafeProposeTransaction
		curve = CurveSecp256k1ECDSAEthereum
	case CurveSecp256k1ECDSAMVM:
		action = ActionEthereumSafeProposeTransaction
		curve = CurveSecp256k1ECDSAMVM
	default:
		return nil, fmt.Errorf("invalid chain: %d", chain)
	}
	switch chain {
	case CurveSecp256k1ECDSAEthereum, CurveSecp256k1ECDSAMVM:
		if assetId == "" {
			return nil, fmt.Errorf("invalid asset_id %s for chain %d", assetId, chain)
		}
	}

	extra := []byte{typ}
	extra = append(extra, uuid.FromStringOrNil(head).Bytes()...)
	extra = append(extra, []byte(destination)...)
	op := &types.Operation{
		Id:     operationId,
		Type:   action,
		Curve:  curve,
		Public: publicKey,
		Extra:  extra,
	}
	return op, nil
}
