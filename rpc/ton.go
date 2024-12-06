package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/util"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
)

type TonRpc struct {
	chainInfo *loader.ChainInfo
}

func NewTonRpc(chainInfo *loader.ChainInfo) *TonRpc {
	return &TonRpc{
		chainInfo: chainInfo,
	}
}

func (t *TonRpc) GetClient() (*ton.APIClient, error) {
	if t.chainInfo.Client == nil {
		client := liteclient.NewConnectionPool()
		err := client.AddConnectionsFromConfigUrl(context.Background(), t.chainInfo.RpcEndPoint)
		if err != nil {
			return nil, fmt.Errorf("error connecting to ton %v", err)
		}
		t.chainInfo.Client = ton.NewAPIClient(client).WithRetry()
	}
	return t.chainInfo.Client.(*ton.APIClient), nil
}

func (t *TonRpc) Client() interface{} {
	return t.chainInfo.Client
}

func (t *TonRpc) Backend() int32 {
	return int32(loader.TonBackend)
}

func (t *TonRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	client, err := t.GetClient()
	if err != nil {
		return 0, err
	}
	masterchainInfo, err := client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, err
	}
	return int64(masterchainInfo.SeqNo), nil
}

func (t *TonRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	return false, 0, fmt.Errorf("not implement")
}

func (t *TonRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return nil, fmt.Errorf("not implement")
}

func (t *TonRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	addr, err := address.ParseAddr(ownerAddr)
	if err != nil {
		return nil, err
	}
	client, err := t.GetClient()
	if err != nil {
		return nil, err
	}
	if util.IsNativeAddress(tokenAddr) {
		block, err := client.GetMasterchainInfo(ctx)
		if err != nil {
			return nil, err
		}
		res, err := client.GetAccount(ctx, block, addr)
		if err != nil {
			return nil, err
		}
		return res.State.Balance.Nano(), nil
	}

	minterAddr, err := address.ParseAddr(tokenAddr)
	if err != nil {
		return nil, err
	}

	jettonClient := jetton.NewJettonMasterClient(client, minterAddr)
	walletClient, err := jettonClient.GetJettonWallet(ctx, addr)
	if err != nil {
		return nil, err
	}

	balance, err := walletClient.GetBalance(ctx)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (t *TonRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return t.GetBalance(ctx, ownerAddr, tokenAddr)
}

func (t *TonRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	return nil, fmt.Errorf("not implement")
}

func (t *TonRpc) IsAddressValid(addr string) bool {
	res, err := address.ParseAddr(addr)
	return err == nil && res != nil
}

func (t *TonRpc) GetChecksumAddress(addr string) string {
	res, _ := address.ParseAddr(addr)
	return res.String()
}
