package ethereum

import (
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/MixinNetwork/go-safe-sdk/ethereum/abi"
	ga "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const (
	ChainEthereum = 2
	ChainMVM      = 4
	ChainPolygon  = 6

	TransactionConfirmations = 1

	ValuePrecision = 18
	ValueDust      = 100000000000000

	TimeLockMinimum = time.Hour * 1
	TimeLockMaximum = time.Hour * 24 * 365

	operationTypeCall         = 0
	operationTypeDelegateCall = 1

	TypeETHTx       = 1
	TypeERC20Tx     = 2
	TypeMultiSendTx = 3

	EthereumEmptyAddress                        = "0x0000000000000000000000000000000000000000"
	EthereumSafeProxyFactoryAddress             = "0x4e1DCf7AD4e460CfD30791CCC4F9c8a4f820ec67"
	EthereumSafeL2Address                       = "0x29fcB43b46531BcA003ddC8FCB67FFE91900C762"
	EthereumCompatibilityFallbackHandlerAddress = "0xfd0732Dc9E303f09fCEf3a7388Ad10A83459Ec99"
	EthereumMultiSendAddress                    = "0x38869bf66a61cF6bDB996A6aE40D5853Fd43B526"
	EthereumSafeGuardAddress                    = "0x5cFE246E7F5d2473e34D0F2B33552076a682d68C"

	predeterminedSaltNonce  = "0xb1073742015cbcf5a3a4d9d1ae33ecf619439710b89475f92e2abd2117e90f90"
	accountContractCode     = "0x608060405234801561001057600080fd5b506040516101e63803806101e68339818101604052602081101561003357600080fd5b8101908080519060200190929190505050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614156100ca576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806101c46022913960400191505060405180910390fd5b806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505060ab806101196000396000f3fe608060405273ffffffffffffffffffffffffffffffffffffffff600054167fa619486e0000000000000000000000000000000000000000000000000000000060003514156050578060005260206000f35b3660008037600080366000845af43d6000803e60008114156070573d6000fd5b3d6000f3fea264697066735822122003d1488ee65e08fa41e58e888a9865554c535f2c77126a82cb4c0f917f31441364736f6c63430007060033496e76616c69642073696e676c65746f6e20616464726573732070726f7669646564"
	safeTxTypehash          = "0xbb8310d486368db6bd6f849402fdd73ad53d316b5a4b2644ad6efe0f941286d8"
	domainSeparatorTypehash = "0x47e79534a245952e8b16893a336b85a3d9ea9fa8c573f3d803afb92a79469218"
	guardStorageSlot        = "0x4a204f620c8c5ccdca3fd54d003badd85ba500436a431f0cbda4f558c93c34c8"
)

type Asset struct {
	Address  string
	Id       string
	Symbol   string
	Name     string
	Decimals uint32
	Chain    byte
}

type Transfer struct {
	Hash         string
	Index        int64
	TokenAddress string
	AssetId      string
	Sender       string
	Receiver     string
	Value        *big.Int
}

func GenerateAssetId(chain byte, assetKey string) string {
	assetKey = strings.ToLower(assetKey)
	err := VerifyAssetKey(assetKey)
	if err != nil {
		panic(assetKey)
	}

	base := GetMixinChainID(int64(chain))
	return BuildChainAssetId(base, assetKey)
}

func VerifyAssetKey(assetKey string) error {
	if len(assetKey) != 42 {
		return fmt.Errorf("invalid mvm asset key %s", assetKey)
	}
	if !strings.HasPrefix(assetKey, "0x") {
		return fmt.Errorf("invalid mvm asset key %s", assetKey)
	}
	if assetKey != strings.ToLower(assetKey) {
		return fmt.Errorf("invalid mvm asset key %s", assetKey)
	}
	k, err := hex.DecodeString(assetKey[2:])
	if err != nil {
		return fmt.Errorf("invalid mvm asset key %s %s", assetKey, err.Error())
	}
	if len(k) != 20 {
		return fmt.Errorf("invalid mvm asset key %s", assetKey)
	}
	return nil
}

func BuildChainAssetId(base, asset string) string {
	h := md5.New()
	io.WriteString(h, base)
	io.WriteString(h, asset)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(sum)
	if err != nil {
		panic(hex.EncodeToString(sum))
	}
	return id.String()
}

func HashMessageForSignature(msg string) ([]byte, error) {
	b, err := hex.DecodeString(msg)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(b), b)))
	return hash.Bytes(), nil
}

func ParseAmount(amount string, decimals int32) *big.Int {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		panic(amount)
	}
	amt = amt.Mul(decimal.New(1, decimals))
	if !amt.IsInteger() {
		panic(amount)
	}
	return amt.BigInt()
}

func UnitAmount(amount *big.Int, decimals int32) string {
	amt := decimal.NewFromBigInt(amount, 0)
	amt = amt.Div(decimal.New(1, decimals))
	return amt.String()
}

func GetEvmChainID(chain int64) int64 {
	switch chain {
	case ChainEthereum:
		return 1
	case ChainPolygon:
		return 137
	case ChainMVM:
		return 73927
	default:
		panic(chain)
	}
}

func GetMixinChainID(chain int64) string {
	switch chain {
	case ChainEthereum:
		return "43d61dcd-e413-450d-80b8-101d5e903357"
	case ChainPolygon:
		return "b7938396-3f94-4e0a-9179-d3440718156f"
	case ChainMVM:
		return "a0ffd769-5850-4b48-9651-d2ae44a3e64d"
	default:
		panic(chain)
	}
}

func FetchAsset(chain byte, rpc, address string) (*Asset, error) {
	addr := common.HexToAddress(address)
	assetId := GenerateAssetId(chain, address)

	conn, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	token, err := abi.NewAsset(addr, conn)
	if err != nil {
		return nil, err
	}
	name, err := token.Name(nil)
	if err != nil {
		return nil, err
	}
	symbol, err := token.Symbol(nil)
	if err != nil {
		return nil, err
	}
	decimals, err := token.Decimals(nil)
	if err != nil {
		return nil, err
	}

	return &Asset{
		Address:  address,
		Id:       assetId,
		Name:     name,
		Symbol:   symbol,
		Decimals: uint32(decimals),
		Chain:    chain,
	}, nil
}

func NormalizeAddress(addr string) string {
	norm := common.HexToAddress(addr).Hex()
	if norm == EthereumEmptyAddress || !strings.EqualFold(norm, addr) {
		return ""
	}
	return norm
}

func PrivToAddress(priv string) (*common.Address, error) {
	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	addr := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &addr, nil
}

func packSafeTransactionArguments(tx *SafeTransaction) []byte {
	bytes32Ty, err := ga.NewType("bytes32", "", nil)
	if err != nil {
		panic(err)
	}
	addressTy, err := ga.NewType("address", "", nil)
	if err != nil {
		panic(err)
	}
	uint256Ty, err := ga.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	arguments := ga.Arguments{
		{
			Type: bytes32Ty,
		},
		{
			Type: addressTy,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: bytes32Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: addressTy,
		},
		{
			Type: addressTy,
		},
		{
			Type: uint256Ty,
		},
	}

	bSafeTxTypehash, err := hex.DecodeString(safeTxTypehash[2:])
	if err != nil {
		panic(err)
	}
	args, err := arguments.Pack(
		toBytes32(bSafeTxTypehash),
		tx.Destination,
		tx.Value,
		toBytes32(crypto.Keccak256(tx.Data)),
		new(big.Int).SetInt64(int64(tx.Operation)),
		tx.SafeTxGas,
		tx.BaseGas,
		tx.GasPrice,
		tx.GasToken,
		tx.RefundReceiver,
		tx.Nonce,
	)
	if err != nil {
		panic(err)
	}
	return args
}

func packDomainSeparatorArguments(chainID int64, safeAddress string) []byte {
	bytes32Ty, err := ga.NewType("bytes32", "", nil)
	if err != nil {
		panic(err)
	}
	addressTy, err := ga.NewType("address", "", nil)
	if err != nil {
		panic(err)
	}
	uint256Ty, err := ga.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	arguments := ga.Arguments{
		{
			Type: bytes32Ty,
		},
		{
			Type: uint256Ty,
		},
		{
			Type: addressTy,
		},
	}

	bDomainSeparatorTypehash, err := hex.DecodeString(domainSeparatorTypehash[2:])
	if err != nil {
		panic(err)
	}
	args, err := arguments.Pack(
		toBytes32(bDomainSeparatorTypehash),
		new(big.Int).SetInt64(chainID),
		common.HexToAddress(safeAddress),
	)
	if err != nil {
		panic(err)
	}
	return args
}

func guardInit(rpc string) (*ethclient.Client, *abi.MixinSafeGuard, error) {
	conn, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, nil, err
	}

	abi, err := abi.NewMixinSafeGuard(common.HexToAddress(EthereumSafeGuardAddress), conn)
	if err != nil {
		return nil, nil, err
	}

	return conn, abi, nil
}

func safeInit(rpc, address string) (*ethclient.Client, *abi.GnosisSafe, error) {
	conn, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, nil, err
	}

	abi, err := abi.NewGnosisSafe(common.HexToAddress(address), conn)
	if err != nil {
		return nil, nil, err
	}

	return conn, abi, nil
}

func toBytes32(b []byte) [32]byte {
	var b32 [32]byte
	copy(b32[:], b[:32])
	return b32
}
