package bitcoin

import (
	"fmt"

	"github.com/MixinNetwork/go-safe-sdk/common"
	"github.com/btcsuite/btcd/txscript"
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

func ExtractPkScriptAddr(pkScript []byte, chain byte) (string, error) {
	cf, err := common.NetConfig(chain)
	if err != nil {
		return "", err
	}
	cls, addrs, threshold, err := txscript.ExtractPkScriptAddrs(pkScript, cf)
	if err != nil {
		return "", err
	}
	if threshold != 1 || len(addrs) != 1 || cls == txscript.NonStandardTy {
		return "", fmt.Errorf("unsupported pkscript %d %v %d", cls, addrs, threshold)
	}
	return addrs[0].EncodeAddress(), nil
}
