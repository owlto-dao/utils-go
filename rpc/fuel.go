package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/machinebox/graphql"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/util"
	"github.com/sentioxyz/fuel-go"
	"github.com/sentioxyz/fuel-go/types"
)

type FuelRpc struct {
	chainInfo     *loader.ChainInfo
	graphqlClient *graphql.Client
}

func NewFuelRpc(chainInfo *loader.ChainInfo) *FuelRpc {
	return &FuelRpc{
		chainInfo:     chainInfo,
		graphqlClient: graphql.NewClient(chainInfo.RpcEndPoint),
	}
}

func (f *FuelRpc) GetClient() *fuel.Client {
	return f.chainInfo.Client.(*fuel.Client)
}

func (f *FuelRpc) Client() interface{} {
	return f.chainInfo.Client
}

func (f *FuelRpc) Backend() int32 {
	return int32(loader.FuelBackend)
}

func (f *FuelRpc) IsAddressValid(addr string) bool {
	return strings.HasPrefix(addr, "0x") && len(addr) == 66 && util.IsHex(addr[2:])
}

func (f *FuelRpc) GetChecksumAddress(addr string) string {
	caddr, _ := util.GetFuelChecksumAddress(addr)
	return caddr
}

func (f *FuelRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	blockNumber, err := f.GetClient().GetLatestBlockHeight(ctx)
	if err != nil {
		return 0, err
	}
	return int64(blockNumber), nil
}

func (f *FuelRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	txn, err := f.GetClient().GetTransaction(ctx, types.QueryTransactionParams{
		Id: types.TransactionId{Hash: common.HexToHash(hash)},
	}, fuel.GetTransactionOption{
		WithReceipts: true,
		WithStatus:   true,
	})
	if err != nil {
		return false, 0, err
	}
	if txn == nil {
		return false, 0, fmt.Errorf("not found fuel tx: %v", hash)
	}
	if txn.Status.SuccessStatus != nil {
		return true, int64(txn.Status.SuccessStatus.BlockHeight), nil
	} else if txn.Status.FailureStatus != nil {
		return false, int64(txn.Status.FailureStatus.BlockHeight), nil
	} else {
		return false, 0, fmt.Errorf("fuel tx: %v is pending", hash)
	}
}

func (f *FuelRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return nil, fmt.Errorf("not implement")
}

func (f *FuelRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	query := `
    query ($owner: Address!, $assetId: AssetId!) {
        balance(owner: $owner, assetId: $assetId) {
            owner
            amount
            assetId
        }
    }
    `
	variables := map[string]interface{}{
		"owner":   ownerAddr,
		"assetId": tokenAddr,
	}

	req := graphql.NewRequest(query)

	for key, value := range variables {
		req.Var(key, value)
	}

	var respData struct {
		Balance struct {
			Owner   string `json:"owner"`
			Amount  string `json:"amount"`
			AssetId string `json:"assetId"`
		} `json:"balance"`
	}

	if err := f.graphqlClient.Run(ctx, req, &respData); err != nil {
		return nil, fmt.Errorf("get balance err: %v", err)
	}

	amountStr := respData.Balance.Amount
	amountBigInt, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	return amountBigInt, nil
}

func (f *FuelRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return f.GetBalance(ctx, ownerAddr, tokenAddr)
}

func (f *FuelRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	return nil, fmt.Errorf("not implement")
}
