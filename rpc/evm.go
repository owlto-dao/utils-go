package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/owlto-dao/utils-go/abi/erc20"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/log"
	"github.com/owlto-dao/utils-go/owlconsts"
	"github.com/owlto-dao/utils-go/pointer"
	"github.com/owlto-dao/utils-go/util"
	"github.com/zksync-sdk/zksync2-go/clients"
	"golang.org/x/crypto/sha3"
)

type EvmRpc struct {
	tokenInfoMgr *loader.TokenInfoManager
	chainInfo    *loader.ChainInfo
	erc20ABI     abi.ABI
	client       EVMClient
}

func NewEvmRpc(chainInfo *loader.ChainInfo) *EvmRpc {
	erc20ABI, _ := abi.JSON(strings.NewReader(erc20.Erc20MetaData.ABI))
	var client EVMClient
	switch chainInfo.Name {
	case owlconsts.ZKSyncEra:
		client = &ZkSyncClientWrapper{client: chainInfo.Client.(*clients.Client)}
	default:
		client = &EthClientWrapper{client: chainInfo.Client.(*ethclient.Client)}
	}

	return &EvmRpc{
		chainInfo:    chainInfo,
		tokenInfoMgr: loader.NewTokenInfoManager(nil, nil),
		erc20ABI:     erc20ABI,
		client:       client,
	}
}

type EVMClient interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
	Client() *rpc.Client
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error)
}

type EthClientWrapper struct {
	client *ethclient.Client
}

func (e *EthClientWrapper) HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error) {
	return e.client.HeaderByNumber(ctx, number)
}

func (e *EthClientWrapper) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return e.client.SuggestGasTipCap(ctx)
}

func (e *EthClientWrapper) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return e.client.SuggestGasPrice(ctx)
}

func (e *EthClientWrapper) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return e.client.BalanceAt(ctx, account, blockNumber)
}

func (e *EthClientWrapper) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	return e.client.TransactionReceipt(ctx, txHash)
}

func (e *EthClientWrapper) BlockNumber(ctx context.Context) (uint64, error) {
	return e.client.BlockNumber(ctx)
}

func (e *EthClientWrapper) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	return e.client.Client().BatchCallContext(ctx, b)
}

func (e *EthClientWrapper) Client() *rpc.Client {
	return e.client.Client()
}

func (e *EthClientWrapper) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return e.client.EstimateGas(ctx, msg)
}

func (e *EthClientWrapper) RawClient() *ethclient.Client {
	return e.client
}

type ZkSyncClientWrapper struct {
	client *clients.Client
}

func (z *ZkSyncClientWrapper) HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error) {
	return z.client.HeaderByNumber(ctx, number)
}

func (z *ZkSyncClientWrapper) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return z.client.SuggestGasTipCap(ctx)
}

func (z *ZkSyncClientWrapper) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return z.client.SuggestGasPrice(ctx)
}

func (z *ZkSyncClientWrapper) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return z.client.BalanceAt(ctx, account, blockNumber)
}

func (z *ZkSyncClientWrapper) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	res, err := z.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}
	return &res.Receipt, err
}

func (z *ZkSyncClientWrapper) BlockNumber(ctx context.Context) (uint64, error) {
	return z.client.BlockNumber(ctx)
}

func (z *ZkSyncClientWrapper) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	return z.client.Client().BatchCallContext(ctx, b)
}

func (z *ZkSyncClientWrapper) Client() *rpc.Client {
	return z.client.Client()
}

func (z *ZkSyncClientWrapper) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return z.client.EstimateGas(ctx, msg)
}

func (z *ZkSyncClientWrapper) RawClient() *clients.Client {
	return z.client
}

func (w *EvmRpc) IsAddressValid(addr string) bool {
	return common.IsHexAddress(addr)
}

func (w *EvmRpc) GetChecksumAddress(addr string) string {
	return common.HexToAddress(addr).Hex()
}

func (w *EvmRpc) GetClient() *ethclient.Client {
	return w.chainInfo.Client.(*ethclient.Client)
}

func (w *EvmRpc) GetZKSyncEraClient() *clients.Client {
	return w.chainInfo.Client.(*clients.Client)
}

func (w *EvmRpc) Client() interface{} {
	return w.chainInfo.Client
}

func (w *EvmRpc) Backend() int32 {
	return 1
}

func (w *EvmRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	if util.IsHexStringZero(tokenAddr) {
		return &loader.TokenInfo{
			TokenName:    w.chainInfo.GasTokenName,
			ChainName:    w.chainInfo.Name,
			TokenAddress: tokenAddr,
			Decimals:     w.chainInfo.GasTokenDecimal,
			FullName:     w.chainInfo.AliasName,
			TotalSupply:  big.NewInt(0),
			Url:          w.chainInfo.ExplorerUrl,
		}, nil
	}
	tokenInfo, ok := w.tokenInfoMgr.GetByChainNameTokenAddr(w.chainInfo.Name, tokenAddr)
	if ok {
		return tokenInfo, nil
	}

	var symbolHex hexutil.Bytes
	var nameHex hexutil.Bytes
	var decimalsHex hexutil.Bytes
	var totalSupplyHex hexutil.Bytes

	symbolData, _ := w.erc20ABI.Pack("symbol")
	decimalsData, _ := w.erc20ABI.Pack("decimals")
	nameData, _ := w.erc20ABI.Pack("name")
	totalSupplyData, _ := w.erc20ABI.Pack("totalSupply")

	var be []rpc.BatchElem
	be = append(be, rpc.BatchElem{
		Method: "eth_call",
		Args: []interface{}{
			map[string]interface{}{
				"to":   tokenAddr,
				"data": hexutil.Encode(symbolData),
			},
			"latest",
		},
		Result: &symbolHex,
	})
	be = append(be, rpc.BatchElem{
		Method: "eth_call",
		Args: []interface{}{
			map[string]interface{}{
				"to":   tokenAddr,
				"data": hexutil.Encode(decimalsData),
			},
			"latest",
		},
		Result: &decimalsHex,
	})
	be = append(be, rpc.BatchElem{
		Method: "eth_call",
		Args: []interface{}{
			map[string]interface{}{
				"to":   tokenAddr,
				"data": hexutil.Encode(nameData),
			},
			"latest"},
		Result: &nameHex,
	})
	be = append(be, rpc.BatchElem{
		Method: "eth_call",
		Args: []interface{}{
			map[string]interface{}{
				"to":   tokenAddr,
				"data": hexutil.Encode(totalSupplyData),
			},
			"latest",
		},
		Result: &totalSupplyHex,
	})

	if err := w.client.BatchCallContext(ctx, be); err != nil {
		return nil, err
	}
	for _, b := range be {
		if b.Error != nil {
			return nil, fmt.Errorf("get token error %s %w", b.Method, b.Error)
		}
	}

	symbol, err := hexutil.Decode(symbolHex.String())
	if err != nil {
		return nil, err
	}

	name, err := hexutil.Decode(nameHex.String())
	if err != nil {
		return nil, err
	}

	decimalsBytes, err := hexutil.Decode(decimalsHex.String())
	if err != nil {
		return nil, err
	}
	decimals := new(big.Int).SetBytes(decimalsBytes)

	totalSupplyBytes, err := hexutil.Decode(totalSupplyHex.String())
	if err != nil {
		return nil, err
	}
	totalSupply := new(big.Int).SetBytes(totalSupplyBytes)

	if decimals.Cmp(common.Big0) <= 0 || len(symbol) == 0 {
		return nil, fmt.Errorf("not found")
	}

	ti := &loader.TokenInfo{
		TokenName:    string(symbol),
		ChainName:    w.chainInfo.Name,
		TokenAddress: tokenAddr,
		Decimals:     int32(decimals.Uint64()),
		FullName:     string(name),
		TotalSupply:  totalSupply,
	}
	w.tokenInfoMgr.AddTokenInfo(ti)
	return ti, nil
}

func (w *EvmRpc) getERC20Contract(tokenAddr string) (*erc20.Erc20, error) {
	if w.chainInfo.Name == owlconsts.ZKSyncEra {
		return erc20.NewErc20(common.HexToAddress(tokenAddr), w.client.(*ZkSyncClientWrapper).RawClient())
	}
	return erc20.NewErc20(common.HexToAddress(tokenAddr), w.client.(*EthClientWrapper).RawClient())
}

func (w *EvmRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	econtract, err := w.getERC20Contract(tokenAddr)
	if err != nil {
		return nil, err
	}
	allowance, err := econtract.Allowance(nil, common.HexToAddress(ownerAddr), common.HexToAddress(spenderAddr))
	if err != nil {
		return nil, err
	}
	return allowance, nil
}

func (w *EvmRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	blockNum := big.NewInt(blockNumber)

	if util.IsHexStringZero(tokenAddr) {
		return w.client.BalanceAt(ctx, common.HexToAddress(ownerAddr), blockNum)
	}

	econtract, err := w.getERC20Contract(tokenAddr)
	if err != nil {
		return nil, err
	}
	return econtract.BalanceOf(&bind.CallOpts{
		Pending:     false,
		Context:     ctx,
		BlockNumber: blockNum,
	}, common.HexToAddress(ownerAddr))
}

func (w *EvmRpc) GetBalance(ctx context.Context, ownerAddr, tokenAddr string) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)

	if util.IsHexStringZero(tokenAddr) {
		return w.client.BalanceAt(ctx, common.HexToAddress(ownerAddr), nil)
	}

	econtract, err := w.getERC20Contract(tokenAddr)
	if err != nil {
		return nil, err
	}
	return econtract.BalanceOf(&bind.CallOpts{Context: ctx}, common.HexToAddress(ownerAddr))
}

func (w *EvmRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	receipt, err := w.client.TransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		return false, 0, err
	}
	if receipt == nil {
		return false, 0, fmt.Errorf("get receipt failed")
	}
	return receipt.Status == ethtypes.ReceiptStatusSuccessful, receipt.BlockNumber.Int64(), nil
}

func (w *EvmRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	blockNumber, err := w.client.BlockNumber(ctx)
	if err != nil {
		log.Errorf("%v get latest block number error %v", w.chainInfo.Name, err)
		return 0, err
	}
	return int64(blockNumber), nil
}

func (w *EvmRpc) EstimateGas(ctx context.Context, fromAddress string, recipient string, tokenAddress string, value *big.Int) (uint64, error) {
	var msg ethereum.CallMsg
	if util.IsNativeAddress(tokenAddress) {
		switch w.chainInfo.Name {
		case owlconsts.Scroll, owlconsts.Ethereum, owlconsts.Optimism, owlconsts.Base, owlconsts.Manta, owlconsts.Linea,
			owlconsts.Bevm, owlconsts.Bevm2, owlconsts.Taiko, owlconsts.AILayer:
			return 21000, nil
		}
		msg = ethereum.CallMsg{
			From:  common.HexToAddress(fromAddress),
			To:    pointer.Ptr(common.HexToAddress(recipient)),
			Value: value,
		}
	} else {
		data := GetERC20TransferData(recipient, value)

		msg = ethereum.CallMsg{
			From: common.HexToAddress(fromAddress),
			To:   pointer.Ptr(common.HexToAddress(tokenAddress)),
			Data: data,
		}
	}

	gasLimit, err := w.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}
	return gasLimit, nil
}

func (w *EvmRpc) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := w.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	return gasPrice, nil
}

func (w *EvmRpc) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	gasTipCap, err := w.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	return gasTipCap, nil
}

func (w *EvmRpc) GetBaseFee(ctx context.Context) (*big.Int, error) {
	header, err := w.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	return header.BaseFee, nil
}

func GetERC20TransferData(recipient string, value *big.Int) []byte {
	transferFnSignature := []byte("transfer(address,uint256)")

	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)

	methID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(common.HexToAddress(recipient).Bytes(), 32)

	var data []byte
	data = append(data, methID...)
	data = append(data, paddedAddress...)
	data = append(data, common.LeftPadBytes(value.Bytes(), 32)...)
	return data
}
