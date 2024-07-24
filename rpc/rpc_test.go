package rpc

import (
	"context"
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/owlto-dao/utils-go/loader"
)

func TestSol(t *testing.T) {
	t.Log("test sol...")
	solRpc := NewSolanaRpc(&loader.ChainInfo{Name: "SolanaMainnet", Client: rpc.New("https://api.mainnet-beta.solana.com")})
	t.Log(solRpc.GetTokenInfo(context.TODO(), "ZyABVe2KLmc1GeGobLkJSjr3U5FPRJVymLX67cojUBm"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zzsReZFpYxg1xYBQbfRKHytGYFEHpPPUCa4NtrHp5pE"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zzMSBu58juvqZbYnqhVMdSFwguiw8oL17T4q3dMWGaN"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "J8qZijXxrypJin5Y27qcTvNjmd5ybF44NJdDKCSkXxWv"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "Fm1hguSMcAcVQ7gLMkyihnUJ5JfcTrBNSz1T4CFFpump"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "J5tzd1ww1V1qrgDUQHVCGqpmpbnEnjzGs9LAqJxwkNde"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zxTtD4MMnEAgHMvXmfgPCyMY61ivxX5zwu12hTSqLoA"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zZRRHGndBuUsbn4VM47RuagdYt57hBbskQ2Ba6K5775"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"))

}
