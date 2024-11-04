package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	_ "github.com/gagliardetto/solana-go"
	"github.com/ninja0404/go-unisat"
	"github.com/owlto-dao/utils-go/apollosdk"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/util"
)

type BitcoinRpc struct {
	chainInfo *loader.ChainInfo
	apolloSDK *apollosdk.ApolloSDK
}

type UnisatAPIConfig map[string]*chainServerBearer

type chainServerBearer struct {
	Server        string
	MemPoolServer string
	Bearer        string
	Timeout       int64 // second
	QueryInterval int64 // second
}

func ParseUnisatAPIConfig(value string) (UnisatAPIConfig, error) {
	var u UnisatAPIConfig
	err := json.Unmarshal([]byte(value), &u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func NewBitcoinRpc(chainInfo *loader.ChainInfo, apolloSDK *apollosdk.ApolloSDK) *BitcoinRpc {
	return &BitcoinRpc{
		chainInfo: chainInfo,
		apolloSDK: apolloSDK,
	}
}

func (w *BitcoinRpc) IsAddressValid(addr string) bool {
	netParams := &chaincfg.MainNetParams
	if w.chainInfo.IsTestnet == 1 {
		netParams = &chaincfg.TestNet3Params
	}
	_, err := btcutil.DecodeAddress(addr, netParams)
	return err == nil
}

func (w *BitcoinRpc) GetChecksumAddress(addr string) string {
	netParams := &chaincfg.MainNetParams
	if w.chainInfo.IsTestnet == 1 {
		netParams = &chaincfg.TestNet3Params
	}
	a, _ := btcutil.DecodeAddress(addr, netParams)
	return a.String()
}

func (w *BitcoinRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	return nil, fmt.Errorf("no impl")
}

func (w *BitcoinRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return w.GetBalance(ctx, ownerAddr, tokenAddr)
}

func (w *BitcoinRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)

	unisatAPIConfig, err := apollosdk.GetConfig(w.apolloSDK, "base_config", "unisat_api_config", ParseUnisatAPIConfig)
	if err != nil {
		return nil, err
	}
	if util.IsHexStringZero(tokenAddr) {
		resp, err := unisat.GetAddressBalance(ctx, unisatAPIConfig[w.chainInfo.Name].Server, unisatAPIConfig[w.chainInfo.Name].Bearer, ownerAddr)
		if err != nil {
			return nil, err
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("unisat GetAddressBalance error: %v", resp.Message)
		}
		return resp.Data.Satoshi, nil
	} else if strings.HasPrefix(tokenAddr, "brc20_") && len(tokenAddr) > 6 {
		brc20 := tokenAddr[6:]
		resp, err := unisat.GetAddressBrc20TickInfo(ctx, unisatAPIConfig[w.chainInfo.Name].Server, unisatAPIConfig[w.chainInfo.Name].Bearer, ownerAddr, brc20)
		if err != nil {
			return nil, err
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("unisat GetAddressBrc20TickInfo error: %v", resp.Message)
		}
		balance, ok := big.NewInt(0).SetString(resp.Data.OverallBalance, 10)
		if ok {
			return balance, nil
		}
		return big.NewInt(0), nil
	} else {
		return big.NewInt(0), fmt.Errorf("not impl")
	}
}

func (w *BitcoinRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return big.NewInt(0), fmt.Errorf("not impl")
}

func (w *BitcoinRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	return false, 0, fmt.Errorf("not impl")
}

func (w *BitcoinRpc) Client() interface{} {
	return w.chainInfo.Client
}

func (w *BitcoinRpc) Backend() int32 {
	return 4
}

func (w *BitcoinRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not impl")
}

//type unisatServer struct {
//	RpcEndPoint string
//	Bearer      string
//}
//
//var chainServerMap = map[string]unisatServer{
//	owlconsts.Bitcoin: {
//		RpcEndPoint: "https://open-api.unisat.io",
//		Bearer:      "4ec6c2086742105639c4716a57b1234144e3bbfd0727ee7692e1cfeaca964105",
//	},
//	owlconsts.FractalBitcoin: {
//		RpcEndPoint: "https://open-api-fractal.unisat.io",
//		Bearer:      "d44b54a14f41891e4b850b9d6aace056911f553a0357299620df4cb6d383af12",
//	},
//	owlconsts.FractalBitcoinTest: {
//		RpcEndPoint: "https://open-api-fractal-testnet.unisat.io",
//		Bearer:      "523c5848192152b2eb1dd20ee08128aa77b9d673812f9fbfb5eb7218518ef195",
//	},
//}
