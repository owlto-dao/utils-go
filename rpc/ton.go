package rpc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/wallet"

	"github.com/owlto-dao/utils-go/loader"
)

const (
	ConfigURLTestnet = "https://tonutils.com/testnet-global.config.json"
	ConfigURLMainnet = "https://tonutils.com/global.config.json"
)

func NewTonAPI(configURL string) ton.APIClientWrapped {
	client := liteclient.NewConnectionPool()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.AddConnectionsFromConfigUrl(ctx, configURL)
	if err != nil {
		panic(err)
	}

	return ton.NewAPIClient(client).WithRetry()
}

type TonAPI interface {
	wallet.TonAPI
	jetton.TonApi
}

type TonRpc struct {
	api TonAPI
}

var _ Rpc = (*TonRpc)(nil)

func NewTonRpc(chainInfo *loader.ChainInfo) *TonRpc {
	var url string
	if chainInfo.IsTestnet == 1 {
		url = ConfigURLTestnet
	} else {
		url = ConfigURLMainnet
	}

	return &TonRpc{api: NewTonAPI(url)}
}

func (w *TonRpc) Client() interface{} {
	return w.api
}

func (w *TonRpc) Backend() int32 {
	return int32(loader.TonBackend)
}

func (w *TonRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	info, err := w.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, err
	}
	return int64(info.SeqNo), nil
}

func (w *TonRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	return false, 0, fmt.Errorf("not impl")
}

func (w *TonRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return big.NewInt(0), fmt.Errorf("not impl")
}

func (w *TonRpc) GetBalance(ctx context.Context, strOwnerAddr string, strTokenAddr string) (*big.Int, error) {
	block, err := w.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}

	ownerAddr, err := address.ParseAddr(strOwnerAddr)
	if err != nil {
		return nil, err
	}

	if len(strTokenAddr) == 0 {
		return w.getNativeBalanceAtBlock(ctx, ownerAddr, block)
	}

	tokenAddr, err := address.ParseAddr(strTokenAddr)
	if err != nil {
		return nil, err
	}

	return w.getTokenBalanceAtBlock(ctx, ownerAddr, tokenAddr, block)
}

func (w *TonRpc) getTokenBalanceAtBlock(ctx context.Context, ownerAddr *address.Address, tokenAddr *address.Address, block *ton.BlockIDExt) (*big.Int, error) {
	client := jetton.NewJettonMasterClient(w.api, tokenAddr)
	wallet, err := client.GetJettonWalletAtBlock(ctx, ownerAddr, block)
	if err != nil {
		return nil, err
	}

	return wallet.GetBalance(ctx)
}

func (w *TonRpc) getNativeBalanceAtBlock(ctx context.Context, addr *address.Address, block *ton.BlockIDExt) (*big.Int, error) {
	acc, err := w.api.WaitForBlock(block.SeqNo).GetAccount(ctx, block, addr)
	if err != nil {
		return nil, err
	}

	if !acc.IsActive {
		return big.NewInt(0), nil
	}

	return acc.State.Balance.Nano(), nil
}

func (w *TonRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return w.GetBalance(ctx, ownerAddr, tokenAddr)
}

// TODO only support total supply
func (w *TonRpc) GetTokenInfo(ctx context.Context, strTokenAddr string) (loader.TokenInfo, error) {
	tokenAddr, err := address.ParseAddr(strTokenAddr)
	if err != nil {
		return loader.TokenInfo{}, err
	}

	client := jetton.NewJettonMasterClient(w.api, tokenAddr)
	data, err := client.GetJettonData(ctx)
	if err != nil {
		return loader.TokenInfo{}, err
	}

	return loader.TokenInfo{
		TotalSupply: data.TotalSupply,
	}, nil
}
