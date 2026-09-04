package operation

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProposalsAcrossSupportedChains(t *testing.T) {
	const (
		operationID = "67b91655-5d09-4686-a2e2-6419388e96e4"
		ownerOne    = "8aef8130-aa9c-418a-871d-e920fed2f0e4"
		ownerTwo    = "ce8491f2-3fde-4d2e-a4cc-4fcf707889c3"
		head        = "a229d702-1888-46b6-9141-f875ffe6c566"
		cancelID    = "af36f755-a48a-3408-8a97-092007f9e2d2"
		lockID      = "358c0e9e-8d9c-4e0f-acde-8945a859763a"
		public      = "02aabbcc"
		destination = "destination"
	)
	hash := bytes.Repeat([]byte{0x42}, 32)

	tests := []struct {
		name          string
		chain         byte
		accountAction uint8
		txAction      uint8
		curve         uint8
	}{
		{name: "bitcoin", chain: SafeChainBitcoin, accountAction: ActionBitcoinSafeProposeAccount, txAction: ActionBitcoinSafeProposeTransaction, curve: CurveSecp256k1ECDSABitcoin},
		{name: "litecoin", chain: SafeChainLitecoin, accountAction: ActionBitcoinSafeProposeAccount, txAction: ActionBitcoinSafeProposeTransaction, curve: CurveSecp256k1ECDSALitecoin},
		{name: "ethereum", chain: SafeChainEthereum, accountAction: ActionEthereumSafeProposeAccount, txAction: ActionEthereumSafeProposeTransaction, curve: CurveSecp256k1ECDSAEthereum},
		{name: "mvm", chain: SafeChainMVM, accountAction: ActionEthereumSafeProposeAccount, txAction: ActionEthereumSafeProposeTransaction, curve: CurveSecp256k1ECDSAMVM},
		{name: "polygon", chain: SafeChainPolygon, accountAction: ActionEthereumSafeProposeAccount, txAction: ActionEthereumSafeProposeTransaction, curve: CurveSecp256k1ECDSAPolygon},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := ProposeAccount(operationID, public, []string{ownerOne, ownerTwo}, 2, tt.chain, 24)
			require.NoError(t, err)
			assert.Equal(t, tt.accountAction, account.Type)
			assert.Equal(t, tt.curve, account.Curve)
			assert.Equal(t, operationID, account.Id)
			assert.Equal(t, public, account.Public)
			require.Len(t, account.Extra, 4+32)
			assert.Equal(t, uint16(24), binary.BigEndian.Uint16(account.Extra[:2]))
			assert.Equal(t, byte(2), account.Extra[2])
			assert.Equal(t, byte(2), account.Extra[3])
			assert.Equal(t, uuid.Must(uuid.FromString(ownerOne)).Bytes(), account.Extra[4:20])

			transaction, err := ProposeTransaction(operationID, public, TransactionTypeNormal, head, destination, tt.chain)
			require.NoError(t, err)
			assert.Equal(t, tt.txAction, transaction.Type)
			assert.Equal(t, tt.curve, transaction.Curve)
			assert.Equal(t, byte(TransactionTypeNormal), transaction.Extra[0])
			assert.Equal(t, uuid.Must(uuid.FromString(head)).Bytes(), transaction.Extra[1:17])
			assert.Equal(t, destination, string(transaction.Extra[17:]))

			batch, err := ProposeBatchTransaction(operationID, public, TransactionTypeNormal, head, hash, tt.chain)
			require.NoError(t, err)
			assert.Equal(t, tt.txAction, batch.Type)
			assert.Equal(t, tt.curve, batch.Curve)
			assert.Equal(t, hash, batch.Extra[17:])

			cancel, err := ProposeCancelTransaction(operationID, public, head, destination, tt.chain, cancelID)
			require.NoError(t, err)
			assert.Equal(t, tt.txAction, cancel.Type)
			assert.Equal(t, tt.curve, cancel.Curve)
			assert.Equal(t, byte(TransactionTypeCancel), cancel.Extra[0])
			assert.Equal(t, uuid.Must(uuid.FromString(cancelID)).Bytes(), cancel.Extra[1:17])

			inheritance, err := ProposeInheritanceTransaction(
				operationID, public, TransactionTypeSetInheritance, head,
				destination, tt.chain, "", hex.EncodeToString(hash), 48,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.txAction, inheritance.Type)
			assert.Equal(t, tt.curve, inheritance.Curve)
			assert.Equal(t, hash, inheritance.Extra[1:33])
			assert.Equal(t, uint16(48), binary.BigEndian.Uint16(inheritance.Extra[33:35]))

			remove, err := ProposeInheritanceTransaction(
				operationID, public, TransactionTypeRemoveInheritance, head,
				destination, tt.chain, lockID, "", 0,
			)
			require.NoError(t, err)
			assert.Equal(t, uuid.Must(uuid.FromString(lockID)).Bytes(), remove.Extra[1:17])
		})
	}
}

func TestProposalValidation(t *testing.T) {
	const (
		id          = "67b91655-5d09-4686-a2e2-6419388e96e4"
		validUUID   = "8aef8130-aa9c-418a-871d-e920fed2f0e4"
		public      = "02aabbcc"
		destination = "destination"
	)

	_, err := ProposeAccount(id, public, nil, 1, 99, 1)
	assert.EqualError(t, err, "invalid chain: 99")
	_, err = ProposeAccount(id, public, []string{"invalid"}, 1, SafeChainBitcoin, 1)
	assert.EqualError(t, err, "invalid uuid invalid")

	_, err = ProposeTransaction(id, public, 0, validUUID, destination, 99)
	assert.EqualError(t, err, "invalid chain: 99")
	_, err = ProposeTransaction(id, public, 0, "invalid", destination, SafeChainBitcoin)
	assert.EqualError(t, err, "invalid head uuid invalid")
	_, err = ProposeTransaction(id, public, 0, "", destination, SafeChainBitcoin)
	assert.EqualError(t, err, "invalid head uuid ")

	_, err = ProposeBatchTransaction(id, public, 0, validUUID, nil, 99)
	assert.EqualError(t, err, "invalid chain: 99")
	_, err = ProposeBatchTransaction(id, public, 0, "invalid", nil, SafeChainBitcoin)
	assert.EqualError(t, err, "invalid head uuid invalid")
	_, err = ProposeBatchTransaction(id, public, 0, "", nil, SafeChainBitcoin)
	assert.EqualError(t, err, "invalid head uuid ")

	_, err = ProposeCancelTransaction(id, public, validUUID, destination, 99, validUUID)
	assert.EqualError(t, err, "invalid chain: 99")
	_, err = ProposeCancelTransaction(id, public, validUUID, destination, SafeChainBitcoin, "invalid")
	assert.EqualError(t, err, "invalid cancel uuid invalid")
	_, err = ProposeCancelTransaction(id, public, "invalid", destination, SafeChainBitcoin, validUUID)
	assert.EqualError(t, err, "invalid head uuid invalid")
	_, err = ProposeCancelTransaction(id, public, "", destination, SafeChainBitcoin, validUUID)
	assert.EqualError(t, err, "invalid head uuid ")

	_, err = ProposeInheritanceTransaction(id, public, TransactionTypeSetInheritance, validUUID, destination, 99, "", "00", 1)
	assert.EqualError(t, err, "invalid chain: 99")
	_, err = ProposeInheritanceTransaction(id, public, 99, validUUID, destination, SafeChainBitcoin, "", "", 1)
	assert.EqualError(t, err, "invalid inheritance tx type: 99")
	_, err = ProposeInheritanceTransaction(id, public, TransactionTypeSetInheritance, validUUID, destination, SafeChainBitcoin, "", "not-hex", 1)
	assert.EqualError(t, err, "invalid inheritance hash not-hex")
	_, err = ProposeInheritanceTransaction(id, public, TransactionTypeRemoveInheritance, validUUID, destination, SafeChainBitcoin, "invalid", "", 1)
	assert.EqualError(t, err, "invalid lock uuid invalid")
	_, err = ProposeInheritanceTransaction(id, public, TransactionTypeRemoveInheritance, validUUID, destination, SafeChainBitcoin, "", "", 1)
	assert.EqualError(t, err, "invalid lock uuid ")
	_, err = ProposeInheritanceTransaction(id, public, TransactionTypeSetInheritance, "invalid", destination, SafeChainBitcoin, "", "00", 1)
	assert.EqualError(t, err, "invalid head uuid invalid")
	_, err = ProposeInheritanceTransaction(id, public, TransactionTypeSetInheritance, "", destination, SafeChainBitcoin, "", "00", 1)
	assert.EqualError(t, err, "invalid head uuid ")
}
