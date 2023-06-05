package operation

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/btcutil"
)

const SigHashType = txscript.SigHashAll | txscript.SigHashAnyOneCanPay

func SignSafeTx(rawStr, privateStr string, chain byte) (string, string, error) {
	rawb, _ := hex.DecodeString(rawStr)
	hpsbt, err := UnmarshalPartiallySignedTransaction(rawb)
	if err != nil {
		return "", "", err
	}
	seed, err := hex.DecodeString(privateStr)
	if err != nil {
		return "", "", err
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

	msg := HashMessageForSignature(msgTx.TxHash().String(), chain)
	sig := ecdsa.Sign(holder, msg).Serialize()
	return hex.EncodeToString(raw), base64.RawURLEncoding.EncodeToString(sig), nil
}

type PartiallySignedTransaction struct {
	*psbt.Packet
}

func (raw *PartiallySignedTransaction) Hash() string {
	return raw.UnsignedTx.TxHash().String()
}

func (raw *PartiallySignedTransaction) Marshal() []byte {
	var rawBuffer bytes.Buffer
	err := raw.Serialize(&rawBuffer)
	if err != nil {
		panic(err)
	}
	rb := rawBuffer.Bytes()
	_, err = psbt.NewFromRawBytes(bytes.NewReader(rb), false)
	if err != nil {
		panic(err)
	}
	return rb
}

func UnmarshalPartiallySignedTransaction(b []byte) (*PartiallySignedTransaction, error) {
	pkt, err := psbt.NewFromRawBytes(bytes.NewReader(b), false)
	if err != nil {
		return nil, err
	}
	return &PartiallySignedTransaction{
		Packet: pkt,
	}, nil
}

func (psbt *PartiallySignedTransaction) SigHash(idx int) []byte {
	tx := psbt.UnsignedTx
	pin := psbt.Inputs[idx]
	satoshi := pin.WitnessUtxo.Value
	pof := txscript.NewCannedPrevOutputFetcher(pin.WitnessScript, satoshi)
	tsh := txscript.NewTxSigHashes(tx, pof)
	hash, err := txscript.CalcWitnessSigHash(pin.WitnessScript, tsh, SigHashType, tx, idx, satoshi)
	if err != nil {
		panic(err)
	}
	sigHashes := psbt.Unknowns[0].Value
	if !bytes.Equal(hash, sigHashes[idx*32:idx*32+32]) {
		panic(idx)
	}
	return hash
}

func HashMessageForSignature(msg string, chain byte) []byte {
	var buf bytes.Buffer
	prefix := "Bitcoin Signed Message:\n"
	switch chain {
	case SafeChainBitcoin:
	case SafeChainLitecoin:
		prefix = "Litecoin Signed Message:\n"
	default:
	}
	_ = wire.WriteVarString(&buf, 0, prefix)
	_ = wire.WriteVarString(&buf, 0, msg)
	return chainhash.DoubleHashB(buf.Bytes())
}

func parseBitcoinCompressedPublicKey(public string) (*btcutil.AddressPubKey, error) {
	pub, err := hex.DecodeString(public)
	if err != nil {
		return nil, err
	}
	return btcutil.NewAddressPubKey(pub, netconfig(ChainBitcoin))
}

func VerifySignatureDER(public string, msg, sig []byte) error {
	pub, err := parseBitcoinCompressedPublicKey(public)
	if err != nil {
		return err
	}
	der, err := ecdsa.ParseDERSignature(sig)
	if err != nil {
		return err
	}
	if der.Verify(msg, pub.PubKey()) {
		return nil
	}
	return fmt.Errorf("bitcoin.VerifySignature(%s, %x, %x)", public, msg, sig)
}
