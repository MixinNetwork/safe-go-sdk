package bitcoin

import "github.com/btcsuite/btcd/wire"

func ValueDust(chain byte) int64 {
	switch chain {
	case ChainBitcoin:
		return 1000
	case ChainLitecoin:
		return 10000
	default:
		panic(chain)
	}
}

func protocolVersion(chain byte) uint32 {
	switch chain {
	case ChainBitcoin:
		return wire.ProtocolVersion
	case ChainLitecoin:
		return 70015
	default:
		panic(chain)
	}
}
