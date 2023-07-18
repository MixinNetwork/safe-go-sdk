package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

func ValueDust(chain byte) (int64, error) {
	switch chain {
	case ChainBitcoin:
		return 1000, nil
	case ChainLitecoin:
		return 10000, nil
	default:
		return 0, fmt.Errorf("invalid chain %d", chain)
	}
}

func protocolVersion(chain byte) (uint32, error) {
	switch chain {
	case ChainBitcoin:
		return wire.ProtocolVersion, nil
	case ChainLitecoin:
		return 70015, nil
	default:
		return 0, fmt.Errorf("invalid chain %d", chain)
	}
}
