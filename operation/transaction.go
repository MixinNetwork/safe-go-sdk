package operation

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const SigHashType = txscript.SigHashAll | txscript.SigHashAnyOneCanPay

func SignSafeTx(rawStr, privateStr string, chain byte) (string, error) {
	rawb, err := base64.RawURLEncoding.DecodeString(rawStr)
	if err != nil {
		rawb, err = hex.DecodeString(rawStr)
		if err != nil {
			return "", err
		}
	}

	hpsbt, err := UnmarshalPartiallySignedTransaction(rawb)
	if err != nil {
		return "", err
	}
	seed, err := hex.DecodeString(privateStr)
	if err != nil {
		return "", err
	}
	holder, _ := btcec.PrivKeyFromBytes(seed)

	msgTx := hpsbt.Packet.UnsignedTx
	log.Printf("%#v", msgTx)
	for idx := range msgTx.TxIn {
		hash, err := hpsbt.SigHash(idx)
		if err != nil {
			return "", err
		}
		sig := ecdsa.Sign(holder, hash).Serialize()
		hpsbt.Packet.Inputs[idx].PartialSigs = []*psbt.PartialSig{{
			PubKey:    holder.PubKey().SerializeCompressed(),
			Signature: sig,
		}}
	}
	raw := hpsbt.Marshal()
	return hex.EncodeToString(raw), nil
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

func (psbt *PartiallySignedTransaction) SigHash(idx int) ([]byte, error) {
	tx := psbt.UnsignedTx
	pin := psbt.Inputs[idx]
	satoshi := pin.WitnessUtxo.Value
	pof := txscript.NewCannedPrevOutputFetcher(pin.WitnessScript, satoshi)
	tsh := txscript.NewTxSigHashes(tx, pof)
	hash, err := txscript.CalcWitnessSigHash(pin.WitnessScript, tsh, SigHashType, tx, idx, satoshi)
	if err != nil {
		return nil, err
	}
	return hash, nil
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
	return btcutil.NewAddressPubKey(pub, netConfig(ChainBitcoin))
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

func CheckTransactionPartiallySignedBy(raw, public string) bool {
	b, _ := hex.DecodeString(raw)
	psbt, _ := UnmarshalPartiallySignedTransaction(b)

	for i := range psbt.Inputs {
		pin := psbt.Inputs[i]
		sigs := make(map[string][]byte, 2)
		for _, ps := range pin.PartialSigs {
			pub := hex.EncodeToString(ps.PubKey)
			sigs[pub] = ps.Signature
		}

		if sigs[public] == nil {
			return false
		}
		hash, err := psbt.SigHash(i)
		if err != nil {
			return false
		}
		err = VerifySignatureDER(public, hash, sigs[public])
		if err != nil {
			return false
		}
	}

	return len(psbt.Inputs) > 0
}
