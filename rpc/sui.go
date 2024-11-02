package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	_ "github.com/gagliardetto/solana-go"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/util"
)

type SuiRpc struct {
	tokenInfoMgr *loader.TokenInfoManager
	chainInfo    *loader.ChainInfo
	client       sui.ISuiAPI
}

func NewSuiRpc(chainInfo *loader.ChainInfo) *SuiRpc {
	return &SuiRpc{
		client:       chainInfo.Client.(sui.ISuiAPI),
		chainInfo:    chainInfo,
		tokenInfoMgr: loader.NewTokenInfoManager(nil, nil),
	}
}

func (w *SuiRpc) IsAddressValid(addr string) bool {
	return strings.HasPrefix(addr, "0x") && len(addr) == 66 && util.IsHex(addr[2:])
}

func (w *SuiRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return w.GetBalance(ctx, ownerAddr, tokenAddr)
}

func (w *SuiRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	tokenAddr = strings.TrimSpace(tokenAddr)
	if util.IsHexStringZero(tokenAddr) {
		tokenAddr = "0x2::sui::SUI"
	}

	tokenInfo, ok := w.tokenInfoMgr.GetByChainNameTokenAddr(w.chainInfo.Name, tokenAddr)
	if ok {
		return tokenInfo, nil
	}

	rsp, err := w.client.SuiXGetCoinMetadata(ctx, models.SuiXGetCoinMetadataRequest{
		CoinType: tokenAddr,
	})
	if err != nil {
		return nil, err
	}

	trsp, err := w.client.SuiXGetTotalSupply(ctx, models.SuiXGetTotalSupplyRequest{
		CoinType: tokenAddr,
	})
	if err != nil {
		return nil, err
	}
	totalSupply, ok := big.NewInt(0).SetString(trsp.Value, 0)
	if !ok {
		return nil, fmt.Errorf("sui totalsupply invalid %s", trsp.Value)
	}

	ti := &loader.TokenInfo{
		TokenName:    rsp.Symbol,
		ChainName:    w.chainInfo.Name,
		TokenAddress: tokenAddr,
		Decimals:     int32(rsp.Decimals),
		FullName:     rsp.Name,
		Icon:         rsp.IconUrl,
		TotalSupply:  totalSupply,
	}
	w.tokenInfoMgr.AddTokenInfo(ti)
	return ti, nil
}

func (w *SuiRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)
	if util.IsHexStringZero(tokenAddr) {
		tokenAddr = "0x2::sui::SUI"
	}
	rsp, err := w.client.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
		Owner:    ownerAddr,
		CoinType: tokenAddr,
	})
	if err != nil {
		return nil, err
	}
	num, ok := big.NewInt(0).SetString(rsp.TotalBalance, 0)
	if !ok {
		return nil, fmt.Errorf("sui balance invalid %s", rsp.TotalBalance)
	}
	return num, nil

}

func (w *SuiRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return big.NewInt(0), fmt.Errorf("not impl")
}

func (w *SuiRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	return false, 0, fmt.Errorf("not impl")
}

func (w *SuiRpc) GetClient() sui.ISuiAPI {
	return w.client
}

func (w *SuiRpc) Client() interface{} {
	return w.chainInfo.Client
}

func (w *SuiRpc) Backend() int32 {
	return 9
}

func (w *SuiRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not impl")
}
