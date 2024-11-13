package rpc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/owlto-dao/utils-go/apollosdk"
	"github.com/owlto-dao/utils-go/loader"
)

type Rpc interface {
	Client() interface{}
	Backend() int32
	GetLatestBlockNumber(ctx context.Context) (int64, error)
	IsTxSuccess(ctx context.Context, hash string) (bool, int64, error)
	GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error)
	GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error)
	GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error)
	GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error)
	IsAddressValid(addr string) bool
	GetChecksumAddress(addr string) string
}

func GetRpc(chainInfo *loader.ChainInfo, apolloSDK *apollosdk.ApolloSDK) (Rpc, error) {
	if chainInfo.Backend == 1 {
		return NewEvmRpc(chainInfo), nil
	} else if chainInfo.Backend == 2 {
		return NewStarknetRpc(chainInfo), nil
	} else if chainInfo.Backend == 3 {
		return NewSolanaRpc(chainInfo), nil
	} else if chainInfo.Backend == 4 {
		return NewBitcoinRpc(chainInfo, apolloSDK), nil
	} else if chainInfo.Backend == 5 {
		return NewZksliteRpc(chainInfo), nil
	} else if chainInfo.Backend == 6 {
		return NewTonRpc(chainInfo), nil
	} else if chainInfo.Backend == 8 {
		return NewBenfenRpc(chainInfo), nil
	} else if chainInfo.Backend == 9 {
		return NewSuiRpc(chainInfo), nil
	} else if chainInfo.Backend == 10 {
		return NewFuelRpc(chainInfo), nil
	}
	return nil, fmt.Errorf("unsupport backend %v", chainInfo.Backend)
}
