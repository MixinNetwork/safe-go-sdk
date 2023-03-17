package safe

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/MixinNetwork/mixin/common"
	"github.com/gofrs/uuid"
)

const (
	OperationTypeWrapper     = 0
	OperationTypeKeygenInput = 1
	OperationTypeSignInput   = 2

	OperationTypeKeygenOutput = 11
	OperationTypeSignOutput   = 12

	CurveSecp256k1ECDSABitcoin   = 1
	CurveSecp256k1ECDSAEthereum  = 2
	CurveSecp256k1SchnorrBitcoin = 3
	CurveEdwards25519Default     = 11
	CurveEdwards25519Mixin       = 12
)

func (o *Operation) IdBytes() []byte {
	uid, err := uuid.FromString(o.Id)
	if err != nil {
		panic(err)
	}
	return uid.Bytes()
}

func DecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(s)
	}
	return b
}

type Operation struct {
	Id     string
	Type   uint8
	Curve  uint8
	Public string
	Extra  []byte
}

// TODO compact format for different type
func (o *Operation) Encode() []byte {
	pub := DecodeHex(o.Public)
	enc := common.NewEncoder()
	writeUUID(enc, o.Id)
	writeByte(enc, o.Type)
	writeByte(enc, o.Curve)
	writeBytes(enc, pub)
	writeBytes(enc, o.Extra)
	return enc.Bytes()
}

func (o *Operation) EncodeBase64() string {
	return base64.RawURLEncoding.EncodeToString(o.Encode())
}

func DecodeOperation(b []byte) (*Operation, error) {
	dec := common.NewDecoder(b)
	id, err := readUUID(dec)
	if err != nil {
		return nil, err
	}
	typ, err := dec.ReadByte()
	if err != nil {
		return nil, err
	}
	crv, err := dec.ReadByte()
	if err != nil {
		return nil, err
	}
	pub, err := readBytes(dec)
	if err != nil {
		return nil, err
	}
	extra, err := readBytes(dec)
	if err != nil {
		return nil, err
	}
	return &Operation{
		Type:   typ,
		Id:     id,
		Curve:  crv,
		Public: hex.EncodeToString(pub),
		Extra:  extra,
	}, nil
}

func readBytes(dec *common.Decoder) ([]byte, error) {
	l, err := dec.ReadByte()
	if err != nil {
		return nil, err
	}
	if l == 0 {
		return nil, nil
	}
	b := make([]byte, l)
	err = dec.Read(b)
	return b, err
}

func writeByte(enc *common.Encoder, b byte) {
	err := enc.WriteByte(b)
	if err != nil {
		panic(err)
	}
}

func writeUUID(enc *common.Encoder, id string) {
	uid, err := uuid.FromString(id)
	if err != nil {
		panic(err)
	}
	enc.Write(uid.Bytes())
}

func writeBytes(enc *common.Encoder, b []byte) {
	l := len(b)
	if l > 200 {
		panic(l)
	}
	writeByte(enc, uint8(l))
	enc.Write(b)
}

func readUUID(dec *common.Decoder) (string, error) {
	var b [16]byte
	err := dec.Read(b[:])
	if err != nil {
		return "", err
	}
	id, err := uuid.FromBytes(b[:])
	return id.String(), err
}
