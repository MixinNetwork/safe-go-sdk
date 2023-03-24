package operation

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/MixinNetwork/mixin/common"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func SignTx(rawStr, privateStr string) {
	rawb, _ := hex.DecodeString(rawStr)
	hpsbt, err := UnmarshalPartiallySignedTransaction(rawb)
	seed, err := hex.DecodeString(privateStr)
	if err != nil {
		panic(err)
	}
	holder, _ := btcec.PrivKeyFromBytes(seed)

	msgTx := hpsbt.Packet.UnsignedTx
	log.Printf("%#v", msgTx)
	for idx := range msgTx.TxIn {
		hash := hpsbt.SigHash(idx)
		sig := ecdsa.Sign(holder, hash).Serialize()
		hpsbt.Packet.Inputs[idx].PartialSigs = []*psbt.PartialSig{{
			PubKey:    holder.PubKey().SerializeCompressed(),
			Signature: sig,
		}}
	}
	raw := hpsbt.Marshal()
	fmt.Printf("psbt: %x\n", raw)

	msg := HashMessageForSignature(msgTx.TxHash().String())
	sig := ecdsa.Sign(holder, msg).Serialize()
	fmt.Printf("signature: %s\n", base64.RawURLEncoding.EncodeToString(sig))
}

type PartiallySignedTransaction struct {
	Hash   string
	Fee    int64
	Packet *psbt.Packet
}

func (raw *PartiallySignedTransaction) Marshal() []byte {
	enc := common.NewEncoder()
	hash, err := hex.DecodeString(raw.Hash)
	if err != nil || len(hash) != 32 {
		panic(raw.Hash)
	}

	var rawBuffer bytes.Buffer
	err = raw.Packet.Serialize(&rawBuffer)
	if err != nil {
		panic(err)
	}
	rb := rawBuffer.Bytes()
	_, err = psbt.NewFromRawBytes(bytes.NewReader(rb), false)
	if err != nil {
		panic(err)
	}

	raw.writeBytes(enc, hash)
	raw.writeBytes(enc, rb)
	enc.WriteUint64(uint64(raw.Fee))
	return enc.Bytes()
}

func UnmarshalPartiallySignedTransaction(b []byte) (*PartiallySignedTransaction, error) {
	dec := common.NewDecoder(b)
	hash, err := dec.ReadBytes()
	if err != nil {
		return nil, err
	}
	raw, err := dec.ReadBytes()
	if err != nil {
		return nil, err
	}
	fee, err := dec.ReadUint64()
	if err != nil {
		return nil, err
	}
	pkt, err := psbt.NewFromRawBytes(bytes.NewReader(raw), false)
	if err != nil {
		return nil, err
	}
	pfee, err := pkt.GetTxFee()
	if err != nil {
		return nil, err
	}
	if uint64(pfee) != fee {
		return nil, fmt.Errorf("fee %d %d", fee, pfee)
	}
	if hex.EncodeToString(hash) != pkt.UnsignedTx.TxHash().String() {
		return nil, fmt.Errorf("hash %x %s", hash, pkt.UnsignedTx.TxHash().String())
	}
	return &PartiallySignedTransaction{
		Hash:   hex.EncodeToString(hash),
		Fee:    int64(fee),
		Packet: pkt,
	}, nil
}

func (t *PartiallySignedTransaction) SigHash(idx int) []byte {
	psbt := t.Packet
	tx := psbt.UnsignedTx
	pin := psbt.Inputs[idx]
	satoshi := pin.WitnessUtxo.Value
	pof := txscript.NewCannedPrevOutputFetcher(pin.WitnessScript, satoshi)
	tsh := txscript.NewTxSigHashes(tx, pof)
	hash, err := txscript.CalcWitnessSigHash(pin.WitnessScript, tsh, txscript.SigHashAll, tx, idx, satoshi)
	if err != nil {
		panic(err)
	}
	sigHashes := psbt.Unknowns[0].Value
	if !bytes.Equal(hash, sigHashes[idx*32:idx*32+32]) {
		panic(idx)
	}
	return hash
}

func HashMessageForSignature(msg string) []byte {
	var buf bytes.Buffer
	_ = wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	_ = wire.WriteVarString(&buf, 0, msg)
	return chainhash.DoubleHashB(buf.Bytes())
}

func (raw *PartiallySignedTransaction) writeBytes(enc *common.Encoder, b []byte) {
	enc.WriteInt(len(b))
	enc.Write(b)
}