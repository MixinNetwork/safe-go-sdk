package operation

import (
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/common"
	"github.com/btcsuite/btcd/wire"
)

func ValueDust(chain byte) (int64, error) {
	switch chain {
	case common.ChainBitcoin:
		return 1000, nil
	case common.ChainLitecoin:
		return 10000, nil
	default:
		return 0, fmt.Errorf("invalid chain: %d", chain)
	}
}

func protocolVersion(chain byte) (uint32, error) {
	switch chain {
	case common.ChainBitcoin:
		return wire.ProtocolVersion, nil
	case common.ChainLitecoin:
		return 70015, nil
	default:
		return 0, fmt.Errorf("invalid chain: %d", chain)
	}
}
