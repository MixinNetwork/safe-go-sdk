package common

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
)

const (
	ChainBitcoin  = 1
	ChainLitecoin = 5
)

func init() {
	ltcParams, _ := NetConfig(ChainLitecoin)
	err := chaincfg.Register(ltcParams)
	if err != nil {
		panic(err)
	}
}

func NetConfig(chain byte) (*chaincfg.Params, error) {
	switch chain {
	case ChainBitcoin:
		return &chaincfg.MainNetParams, nil
	case ChainLitecoin:
		return &chaincfg.Params{
			Net:             0xdbb6c0fb,
			Bech32HRPSegwit: "ltc",

			PubKeyHashAddrID:        0x30,
			ScriptHashAddrID:        0x32,
			WitnessPubKeyHashAddrID: 0x06,
			WitnessScriptHashAddrID: 0x0A,

			HDPublicKeyID:  [4]byte{0x01, 0x9d, 0xa4, 0x64},
			HDPrivateKeyID: [4]byte{0x01, 0x9d, 0x9c, 0xfe},
		}, nil
	default:
		return nil, fmt.Errorf("invalid chain %d", chain)
	}
}
