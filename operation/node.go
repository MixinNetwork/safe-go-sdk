package operation

import (
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/types"
	"github.com/MixinNetwork/trusted-group/mtg"
	"github.com/gofrs/uuid/v5"
)

type MixinExtraPack struct {
	T uuid.UUID
	M string `msgpack:",omitempty"`
}

func DecodeExtra(aesKey []byte, memo string) (*types.Operation, error) {
	_, _, m := mtg.DecodeMixinExtra(memo)
	if m == "" {
		return nil, fmt.Errorf("invalid extra format: %s", memo)
	}
	b := AESDecrypt(aesKey, []byte(m))
	return types.DecodeOperation(b)
}
