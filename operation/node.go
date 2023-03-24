package operation

import (
	"encoding/hex"
	"fmt"
	"unicode/utf8"

	"github.com/MixinNetwork/go-safe-sdk/types"
	"github.com/MixinNetwork/trusted-group/mtg"
	"github.com/gofrs/uuid"
)

type MixinExtraPack struct {
	T uuid.UUID
	M string `msgpack:",omitempty"`
}

func DecodeMixinExtra(b []byte) *MixinExtraPack {
	var p MixinExtraPack
	err := mtg.MsgpackUnmarshal(b, &p)
	if err == nil && (p.M != "" || p.T.String() != uuid.Nil.String()) {
		return &p
	}
	if utf8.Valid(b) {
		p.M = string(b)
	} else {
		p.M = hex.EncodeToString(b)
	}
	return &p
}

func DecodeExtra(aesKey []byte, memo string) (*types.Operation, error) {
	memoBuf, err := hex.DecodeString(memo)
	if err != nil {
		return nil, err
	}
	mep := DecodeMixinExtra(memoBuf)
	msp := mtg.DecodeMixinExtra(mep.M)
	if msp == nil {
		return nil, fmt.Errorf("empty memo")
	}
	b := AESDecrypt(aesKey, []byte(msp.M))
	return types.DecodeOperation(b)
}
