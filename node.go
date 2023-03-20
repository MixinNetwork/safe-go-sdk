package safe

import (
	"fmt"

	"github.com/MixinNetwork/trusted-group/mtg"
)

func DecodeExtra(aesKey []byte, memo string) (*Operation, error) {
	msp := mtg.DecodeMixinExtra(memo)
	if msp == nil {
		return nil, fmt.Errorf("empty memo")
	}
	b := AESDecrypt(aesKey, []byte(msp.M))
	return DecodeOperation(b)
}
