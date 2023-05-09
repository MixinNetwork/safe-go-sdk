package bitcoin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	InputTypeP2WPKHAccoutant             = 1
	InputTypeP2WSHMultisigHolderSigner   = 2
	InputTypeP2WSHMultisigObserverSigner = 3

	MaxTransactionSequence = 0xffffffff
	MaxStandardTxWeight    = 300000
)

type Input struct {
	TransactionHash string
	Index           uint32
	Satoshi         int64
	Script          []byte
	Sequence        uint32
	RouteBackup     bool
}

type Output struct {
	Address  string
	Satoshi  int64
	Height   uint64
	Coinbase bool
}

func EstimateTransactionFee(mainInputs []*Input, feeInputs []*Input, outputs []*Output, fvb int64, rid []byte) (int64, error) {
	msgTx := wire.NewMsgTx(2)

	mainAddress, mainSatoshi, err := addInputs(msgTx, mainInputs)
	if err != nil {
		return 0, fmt.Errorf("addInputs(main) => %v", err)
	}
	_, feeSatoshi, err := addInputs(msgTx, feeInputs)
	if err != nil {
		return 0, fmt.Errorf("addInputs(fee) => %v", err)
	}

	var outputSatoshi int64
	for _, out := range outputs {
		err := addOutput(msgTx, out.Address, out.Satoshi)
		if err != nil {
			return 0, fmt.Errorf("addOutput(%s, %d) => %v", out.Address, out.Satoshi, err)
		}
		outputSatoshi = outputSatoshi + out.Satoshi
	}
	if outputSatoshi > mainSatoshi {
		return 0, fmt.Errorf("insufficient %s %d %d", "main", mainSatoshi, outputSatoshi)
	}
	if change := mainSatoshi - outputSatoshi; change > 0 {
		err := addOutput(msgTx, mainAddress, change)
		if err != nil {
			return 0, fmt.Errorf("addOutput(%s, %d) => %v", mainAddress, change, err)
		}
	}

	estvb := (40 + len(msgTx.TxIn)*300 + (len(msgTx.TxOut)+1)*128) / 4
	if len(rid) > 0 && len(rid) <= 64 {
		estvb += len(rid)
	}
	feeConsumed := fvb * int64(estvb)
	if feeConsumed > feeSatoshi {
		return feeConsumed - feeSatoshi, fmt.Errorf("insufficient %s %d %d", "fee", feeConsumed, feeSatoshi)
	}
	return 0, nil
}

func addInputs(tx *wire.MsgTx, inputs []*Input) (string, int64, error) {
	var address string
	var inputSatoshi int64
	for _, input := range inputs {
		addr, err := addInput(tx, input)
		if err != nil {
			return "", 0, err
		}
		if address == "" {
			address = addr
		}
		if address != addr {
			return "", 0, fmt.Errorf("input address %s %s", address, addr)
		}
		inputSatoshi = inputSatoshi + input.Satoshi
	}
	return address, inputSatoshi, nil
}

func addInput(tx *wire.MsgTx, in *Input) (string, error) {
	var addr string
	hash, err := chainhash.NewHashFromStr(in.TransactionHash)
	if err != nil {
		return "", err
	}
	txIn := &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  *hash,
			Index: in.Index,
		},
	}
	typ := checkScriptType(in.Script)
	if in.RouteBackup {
		typ = InputTypeP2WSHMultisigObserverSigner
	}
	switch typ {
	case InputTypeP2WPKHAccoutant:
		in.Script = btcutil.Hash160(in.Script)
		wpkh, err := btcutil.NewAddressWitnessPubKeyHash(in.Script, &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		builder := txscript.NewScriptBuilder()
		builder.AddOp(txscript.OP_0)
		builder.AddData(in.Script)
		script, err := builder.Script()
		if err != nil {
			return "", err
		}
		in.Script = script
		addr = wpkh.EncodeAddress()
		txIn.Sequence = MaxTransactionSequence
	case InputTypeP2WSHMultisigHolderSigner:
		msh := sha256.Sum256(in.Script)
		mwsh, err := btcutil.NewAddressWitnessScriptHash(msh[:], &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		addr = mwsh.EncodeAddress()
		txIn.Sequence = MaxTransactionSequence
	case InputTypeP2WSHMultisigObserverSigner:
		msh := sha256.Sum256(in.Script)
		mwsh, err := btcutil.NewAddressWitnessScriptHash(msh[:], &chaincfg.MainNetParams)
		if err != nil {
			return "", err
		}
		addr = mwsh.EncodeAddress()
		txIn.Sequence = in.Sequence
	default:
		return "", fmt.Errorf("invalid input type %d", typ)
	}
	if txIn.Sequence == 0 {
		return "", fmt.Errorf("invalid sequence %d", in.Sequence)
	}
	tx.AddTxIn(txIn)
	return addr, nil
}

func addOutput(tx *wire.MsgTx, address string, satoshi int64) error {
	addr, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		return err
	}
	script, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return err
	}
	tx.AddTxOut(wire.NewTxOut(satoshi, script))
	return nil
}

func checkScriptType(script []byte) int {
	if len(script) == 33 {
		return InputTypeP2WPKHAccoutant
	}
	if len(script) > 100 {
		return InputTypeP2WSHMultisigHolderSigner
	}
	panic(hex.EncodeToString(script))
}