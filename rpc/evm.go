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
	"golang.org/x/crypto/sha3"
)

type EvmRpc struct {
	tokenInfoMgr *loader.TokenInfoManager
	chainInfo    *loader.ChainInfo
	erc20ABI     abi.ABI
}

func NewEvmRpc(chainInfo *loader.ChainInfo) *EvmRpc {
	erc20ABI, _ := abi.JSON(strings.NewReader(erc20.Erc20ABI))
	return &EvmRpc{
		chainInfo:    chainInfo,
		tokenInfoMgr: loader.NewTokenInfoManager(nil, nil),
		erc20ABI:     erc20ABI,
	}
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

	if err := w.GetClient().Client().BatchCallContext(ctx, be); err != nil {
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

func (w *EvmRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	econtract, err := erc20.NewErc20(common.HexToAddress(tokenAddr), w.GetClient())
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

	if util.IsHexStringZero(tokenAddr) {
		nativeBalance, err := w.GetClient().BalanceAt(ctx, common.HexToAddress(ownerAddr), big.NewInt(blockNumber))
		if err != nil {
			return nil, err
		}
		return nativeBalance, nil
	} else {
		econtract, err := erc20.NewErc20(common.HexToAddress(tokenAddr), w.GetClient())
		if err != nil {
			return nil, err
		}

		balance, err := econtract.BalanceOf(&bind.CallOpts{
			Pending:     false,
			Context:     ctx,
			BlockNumber: big.NewInt(blockNumber),
		}, common.HexToAddress(ownerAddr))

		if err != nil {
			return nil, err
		}
		return balance, nil
	}
}

func (w *EvmRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)

	if util.IsHexStringZero(tokenAddr) {
		nativeBalance, err := w.GetClient().BalanceAt(ctx, common.HexToAddress(ownerAddr), nil)
		if err != nil {
			return nil, err
		}
		return nativeBalance, nil
	} else {
		econtract, err := erc20.NewErc20(common.HexToAddress(tokenAddr), w.GetClient())
		if err != nil {
			return nil, err
		}
		balance, err := econtract.BalanceOf(nil, common.HexToAddress(ownerAddr))

		if err != nil {
			return nil, err
		}
		return balance, nil
	}
}

func (w *EvmRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	receipt, err := w.GetClient().TransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		return false, 0, err
	}
	if receipt == nil {
		return false, 0, fmt.Errorf("get receipt failed")
	}
	return receipt.Status == ethtypes.ReceiptStatusSuccessful, receipt.BlockNumber.Int64(), nil
}

func (w *EvmRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	blockNumber, err := w.GetClient().BlockNumber(ctx)
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
			isContract, err := w.IsContractAddress(ctx, recipient)
			if err != nil {
				return 0, err
			}
			if !isContract {
				return 21000, nil
			}
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

	gasLimit, err := w.GetClient().EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}
	return gasLimit, nil
}

func (w *EvmRpc) IsContractAddress(ctx context.Context, address string) (bool, error) {
	addr := common.HexToAddress(address)

	code, err := w.GetClient().CodeAt(ctx, addr, nil)
	if err != nil {
		return false, err
	}

	return len(code) > 0, nil
}

func (w *EvmRpc) SuggestGasPrice() (*big.Int, error) {
	gasPrice, err := w.GetClient().SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	return gasPrice, nil
}

func (w *EvmRpc) SuggestGasTipCap() (*big.Int, error) {
	gasTipCap, err := w.GetClient().SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, err
	}
	return gasTipCap, nil
}

func (w *EvmRpc) GetBaseFee() (*big.Int, error) {
	header, err := w.GetClient().HeaderByNumber(context.Background(), nil)
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
