package rpc

import (
	"context"
	"testing"

	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/owlto-dao/utils-go/loader"
	"github.com/shopspring/decimal"
)

func TestSol(t *testing.T) {
	t.Log("test evm...")
	t.Log(decimal.NewFromFloat(123).Shift(0).String())
	client, _ := ethclient.Dial("https://bsc-dataseed.bnbchain.org")
	evmRpc := NewEvmRpc(&loader.ChainInfo{Name: "BnbMainnet", Client: client})
	t.Log(evmRpc.GetTokenInfo(context.TODO(), "0x1E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82"))
	t.Log(evmRpc.GetTokenInfo(context.TODO(), "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82"))
	t.Log(evmRpc.GetBalanceAtBlockNumber(context.TODO(), "0xcD98738Cc9F411cD4C001e883c6e69F108A68acd", "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82", 41929855))

	t.Log("test sol...")
	solRpc := NewSolanaRpc(&loader.ChainInfo{Name: "SolanaMainnet", Client: rpc.New("https://api.mainnet-beta.solana.com")})
	t.Log(solRpc.GetTokenInfo(context.TODO(), "5k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zzsReZFpYxg1xYBQbfRKHytGYFEHpPPUCa4NtrHp5pE"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zzMSBu58juvqZbYnqhVMdSFwguiw8oL17T4q3dMWGaN"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "J8qZijXxrypJin5Y27qcTvNjmd5ybF44NJdDKCSkXxWv"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "Fm1hguSMcAcVQ7gLMkyihnUJ5JfcTrBNSz1T4CFFpump"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "J5tzd1ww1V1qrgDUQHVCGqpmpbnEnjzGs9LAqJxwkNde"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zxTtD4MMnEAgHMvXmfgPCyMY61ivxX5zwu12hTSqLoA"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "zZRRHGndBuUsbn4VM47RuagdYt57hBbskQ2Ba6K5775"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"))
	t.Log(solRpc.GetTokenInfo(context.TODO(), "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"))

	suiRpc := NewSuiRpc(&loader.ChainInfo{Name: "SuiMainnet", Client: sui.NewSuiClient("https://fullnode.mainnet.sui.io:443")})
	t.Log(suiRpc.GetTokenInfo(context.TODO(), "0xaf5c10e828852ed8f5cdcc824a80dbe11693be84284aee5dea47d6c2810b4a1::hopcat::HOPCAT"))
}
