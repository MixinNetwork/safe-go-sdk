package ethereum

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetSafeLastTxTime(rpc, address string) (time.Time, error) {
	conn, abi, err := guardInit(rpc)
	if err != nil {
		return time.Time{}, err
	}
	defer conn.Close()

	addr := common.HexToAddress(address)
	timestamp, err := abi.SafeLastTxTime(nil, addr)
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(timestamp.Int64(), 0)
	return t, nil
}

func VerifyHolderKey(public string) error {
	_, err := ParseEthereumCompressedPublicKey(public)
	return err
}

func VerifyMessageSignature(public string, msg, sig []byte) error {
	hash, err := HashMessageForSignature(hex.EncodeToString(msg))
	if err != nil {
		return err
	}
	return VerifyHashSignature(public, hash, sig)
}

func VerifyHashSignature(public string, hash, sig []byte) error {
	pub, err := hex.DecodeString(public)
	if err != nil {
		panic(public)
	}
	signed := crypto.VerifySignature(pub, hash, sig[:64])
	if signed {
		return nil
	}
	return fmt.Errorf("crypto.VerifySignature(%s, %x, %x)", public, hash, sig)
}

func ParseEthereumCompressedPublicKey(public string) (*common.Address, error) {
	pub, err := hex.DecodeString(public)
	if err != nil {
		return nil, err
	}

	publicKey, err := crypto.DecompressPubkey(pub)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*publicKey)
	return &addr, nil
}
