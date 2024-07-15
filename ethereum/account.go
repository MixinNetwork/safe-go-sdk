package ethereum

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetSafeAccountGuard(rpc, address string) (string, error) {
	conn, abi, err := safeInit(rpc, address)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	bGuardOffet, err := hex.DecodeString(guardStorageSlot[2:])
	if err != nil {
		return "", err
	}
	bGuard, err := abi.GetStorageAt(nil, new(big.Int).SetBytes(bGuardOffet), new(big.Int).SetInt64(1))
	if err != nil {
		if strings.Contains(err.Error(), "no contract code at given address") {
			return "", nil
		}
		return "", err
	}
	guardAddress := common.BytesToAddress(bGuard)
	return guardAddress.Hex(), nil
}

func GetSafeLastTxTime(rpc, address string) (time.Time, error) {
	guardAddress, err := GetSafeAccountGuard(rpc, address)
	if err != nil {
		return time.Time{}, err
	}
	switch guardAddress {
	case "", EthereumEmptyAddress:
		panic(fmt.Errorf("safe %s is not deployed or guard is not enabled", address))
	}

	conn, abi, err := guardInit(rpc, guardAddress)
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

func GetOwners(rpc, address string) ([]common.Address, error) {
	conn, abi, err := safeInit(rpc, address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	os, err := abi.GetOwners(nil)
	if err != nil {
		return nil, err
	}
	return os, nil
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

func ParseEthereumUncompressedPublicKey(public string) (*common.Address, error) {
	xPub, _ := hdkeychain.NewKeyFromString(public)
	ecPub, _ := xPub.ECPubKey()
	pub := ecPub.SerializeCompressed()

	publicKey, err := crypto.DecompressPubkey(pub)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*publicKey)
	return &addr, nil
}
