package rpc

import (
	"context"
	"fmt"
	"math/big"

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
	GetTokenInfo(ctx context.Context, tokenAddr string) (loader.TokenInfo, error)
}

func GetRpc(chainInfo *loader.ChainInfo) (Rpc, error) {
	switch chainInfo.Backend {
	case loader.EthereumBackend:
		return NewEvmRpc(chainInfo), nil
	case loader.StarknetBackend:
		return NewStarknetRpc(chainInfo), nil
	case loader.SolanaBackend:
		return NewSolanaRpc(chainInfo), nil
	case loader.BitcoinBackend:
		return NewBitcoinRpc(chainInfo), nil
	case loader.ZksliteBackend:
		return NewZksliteRpc(chainInfo), nil
	case loader.TonBackend:
		return NewTonRpc(chainInfo), nil
	default:
		return nil, fmt.Errorf("unsupport backend %v", chainInfo.Backend)
	}
}
