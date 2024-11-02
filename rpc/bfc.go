package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	_ "github.com/gagliardetto/solana-go"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/owlto-dao/utils-go/network"
	"github.com/owlto-dao/utils-go/util"
)

type BenfenRpc struct {
	chainInfo *loader.ChainInfo
}

func NewBenfenRpc(chainInfo *loader.ChainInfo) *BenfenRpc {
	return &BenfenRpc{
		chainInfo: chainInfo,
	}
}

func (w *BenfenRpc) IsAddressValid(addr string) bool {
	return strings.HasPrefix(addr, "BFC") && len(addr) == 71 && util.IsHex(addr[3:])
}

func (w *BenfenRpc) GetChecksumAddress(addr string) string {
	return addr
}

func (w *BenfenRpc) GetBalanceAtBlockNumber(ctx context.Context, ownerAddr string, tokenAddr string, blockNumber int64) (*big.Int, error) {
	return w.GetBalance(ctx, ownerAddr, tokenAddr)
}

func (w *BenfenRpc) GetTokenInfo(ctx context.Context, tokenAddr string) (*loader.TokenInfo, error) {
	return nil, fmt.Errorf("no impl")
}

func (w *BenfenRpc) GetBalance(ctx context.Context, ownerAddr string, tokenAddr string) (*big.Int, error) {
	ownerAddr = strings.TrimSpace(ownerAddr)
	tokenAddr = strings.TrimSpace(tokenAddr)

	var data map[string]interface{}
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "bfcx_getBalance",
		"params":  []string{ownerAddr, tokenAddr},
	}

	err := network.Request(w.chainInfo.RpcEndPoint, request, &data)
	if err != nil {
		return big.NewInt(0), err
	}

	if result, ok := data["result"].(map[string]interface{}); ok {
		if totalBalance, ok := result["totalBalance"].(string); ok {
			tb, ok := big.NewInt(0).SetString(totalBalance, 10)
			if ok {
				return tb, nil
			}
		}
	}
	return big.NewInt(0), fmt.Errorf("result invalid %v", data)

}

func (w *BenfenRpc) GetAllowance(ctx context.Context, ownerAddr string, tokenAddr string, spenderAddr string) (*big.Int, error) {
	return big.NewInt(0), fmt.Errorf("not impl")
}

func (w *BenfenRpc) IsTxSuccess(ctx context.Context, hash string) (bool, int64, error) {
	return false, 0, fmt.Errorf("not impl")
}

func (w *BenfenRpc) Client() interface{} {
	return w.chainInfo.Client
}

func (w *BenfenRpc) Backend() int32 {
	return 8
}

func (w *BenfenRpc) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not impl")
}
