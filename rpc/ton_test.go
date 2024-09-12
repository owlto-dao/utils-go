package rpc

import (
	"context"
	"testing"

	"github.com/owlto-dao/utils-go/loader"
	"github.com/stretchr/testify/require"
)

func TestTonRpc(t *testing.T) {
	rpc := NewTonRpc(&loader.ChainInfo{IsTestnet: 1})
	ctx := context.Background()

	require.Equal(t, rpc.Backend(), int32(loader.TonBackend))

	height, err := rpc.GetLatestBlockNumber(ctx)
	require.NoError(t, err)
	require.Greater(t, height, int64(0))

	amount, err := rpc.GetBalance(ctx, "0QBTLqrHtXMNGVhiiyR0iwAthVUawzeu7t9EooFVcYLczTXW", "")
	require.NoError(t, err)
	require.Greater(t, amount.Int64(), int64(0))

	amount, err = rpc.GetBalance(ctx, "0QBTLqrHtXMNGVhiiyR0iwAthVUawzeu7t9EooFVcYLczTXW", "kQAiboDEv_qRrcEdrYdwbVLNOXBHwShFbtKGbQVJ2OKxY_Di")
	require.NoError(t, err)
	require.Greater(t, amount.Int64(), int64(0))

	info, err := rpc.GetTokenInfo(ctx, "kQAiboDEv_qRrcEdrYdwbVLNOXBHwShFbtKGbQVJ2OKxY_Di")
	require.NoError(t, err)
	require.Greater(t, info.TotalSupply.Int64(), int64(0))
}
