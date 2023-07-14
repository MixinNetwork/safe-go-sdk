package bitcoin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/MixinNetwork/go-safe-sdk/common"
	"github.com/MixinNetwork/go-safe-sdk/operation"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
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

func EstimateTransactionFee(mainInputs []*Input, feeInputs []*Input, outputs []*Output, fvb int64, rid []byte, chain byte) (int64, error) {
	msgTx := wire.NewMsgTx(2)

	mainAddress, mainSatoshi, err := addInputs(msgTx, mainInputs, chain)
	if err != nil {
		return 0, fmt.Errorf("addInputs(main) => %v", err)
	}
	_, feeSatoshi, err := addInputs(msgTx, feeInputs, chain)
	if err != nil {
		return 0, fmt.Errorf("addInputs(fee) => %v", err)
	}

	var outputSatoshi int64
	for _, out := range outputs {
		err := addOutput(msgTx, out.Address, out.Satoshi, chain)
		if err != nil {
			return 0, fmt.Errorf("addOutput(%s, %d, %d) => %v", out.Address, out.Satoshi, chain, err)
		}
		outputSatoshi = outputSatoshi + out.Satoshi
	}
	if outputSatoshi > mainSatoshi {
		return 0, fmt.Errorf("insufficient %s %d %d", "main", mainSatoshi, outputSatoshi)
	}
	if change := mainSatoshi - outputSatoshi; change > 0 {
		err := addOutput(msgTx, mainAddress, change, chain)
		if err != nil {
			return 0, fmt.Errorf("addOutput(%s, %d, %d) => %v", mainAddress, change, chain, err)
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

func BuildPartiallySignedTransaction(mainInputs []*Input, outputs []*Output, rid []byte, chain byte) (*operation.PartiallySignedTransaction, error) {
	msgTx := wire.NewMsgTx(2)

	mainAddress, mainSatoshi, err := addInputs(msgTx, mainInputs, chain)
	if err != nil {
		return nil, fmt.Errorf("addInputs(main) => %v", err)
	}

	var outputSatoshi int64
	for _, out := range outputs {
		err := addOutput(msgTx, out.Address, out.Satoshi, chain)
		if err != nil {
			return nil, fmt.Errorf("addOutput(%s, %d) => %v", out.Address, out.Satoshi, err)
		}
		outputSatoshi = outputSatoshi + out.Satoshi
	}
	if outputSatoshi > mainSatoshi {
		return nil, fmt.Errorf("insufficient main %d %d", mainSatoshi, outputSatoshi)
	}
	mainChange := mainSatoshi - outputSatoshi
	dust, err := operation.ValueDust(chain)
	if err != nil {
		return nil, err
	}
	if mainChange > dust {
		err := addOutput(msgTx, mainAddress, mainChange, chain)
		if err != nil {
			return nil, fmt.Errorf("addOutput(%s, %d) => %v", mainAddress, mainChange, err)
		}
	}

	estvb := (40 + len(msgTx.TxIn)*300 + (len(msgTx.TxOut)+1)*128) / 4
	if len(rid) > 0 && len(rid) <= 64 {
		estvb += len(rid)
	}

	if len(rid) > 0 && len(rid) <= 64 {
		builder := txscript.NewScriptBuilder()
		builder.AddOp(txscript.OP_RETURN)
		builder.AddData(rid)
		script, err := builder.Script()
		if err != nil {
			return nil, fmt.Errorf("return(%x) => %v", rid, err)
		}
		msgTx.AddTxOut(wire.NewTxOut(0, script))
	}

	rawBytes, err := operation.MarshalWiredTransaction(msgTx, wire.BaseEncoding, chain)
	if err != nil {
		return nil, err
	}
	if len(rawBytes) > estvb {
		return nil, fmt.Errorf("estimation %d %d", len(rawBytes), estvb)
	}
	if estvb*4 > MaxStandardTxWeight {
		return nil, fmt.Errorf("large %d", estvb)
	}

	tx := btcutil.NewTx(msgTx)
	err = blockchain.CheckTransactionSanity(tx)
	if err != nil {
		return nil, fmt.Errorf("blockchain.CheckTransactionSanity() => %v", err)
	}
	lockTime := time.Now().Add(TimeLockMaximum)
	err = mempool.CheckTransactionStandard(tx, txscript.LockTimeThreshold, lockTime, mempool.DefaultMinRelayTxFee, 2)
	if err != nil {
		return nil, fmt.Errorf("mempool.CheckTransactionStandard() => %v", err)
	}

	pkt, err := psbt.NewFromUnsignedTx(msgTx)
	if err != nil {
		return nil, fmt.Errorf("psbt.NewFromUnsignedTx() => %v", err)
	}
	for i, in := range mainInputs {
		address := mainAddress
		addr, err := btcutil.DecodeAddress(address, common.NetConfig(chain))
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		pin := psbt.NewPsbtInput(nil, &wire.TxOut{
			Value:    in.Satoshi,
			PkScript: pkScript,
		})
		pin.WitnessScript = in.Script
		pin.SighashType = SigHashType
		if !pin.IsSane() {
			return nil, fmt.Errorf("!pin.IsSane")
		}
		pkt.Inputs[i] = *pin
	}
	err = pkt.SanityCheck()
	if err != nil {
		return nil, fmt.Errorf("psbt.SanityCheck() => %v", err)
	}

	return &operation.PartiallySignedTransaction{
		Packet: pkt,
	}, nil
}

func addInputs(tx *wire.MsgTx, inputs []*Input, chain byte) (string, int64, error) {
	var address string
	var inputSatoshi int64
	for _, input := range inputs {
		addr, err := addInput(tx, input, chain)
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

func addInput(tx *wire.MsgTx, in *Input, chain byte) (string, error) {
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
		wpkh, err := btcutil.NewAddressWitnessPubKeyHash(in.Script, common.NetConfig(chain))
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
		mwsh, err := btcutil.NewAddressWitnessScriptHash(msh[:], common.NetConfig(chain))
		if err != nil {
			return "", err
		}
		addr = mwsh.EncodeAddress()
		txIn.Sequence = MaxTransactionSequence
	case InputTypeP2WSHMultisigObserverSigner:
		msh := sha256.Sum256(in.Script)
		mwsh, err := btcutil.NewAddressWitnessScriptHash(msh[:], common.NetConfig(chain))
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

func addOutput(tx *wire.MsgTx, address string, satoshi int64, chain byte) error {
	addr, err := btcutil.DecodeAddress(address, common.NetConfig(chain))
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
