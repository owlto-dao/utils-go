package loader

import (
	"database/sql"
	"sync"

	"github.com/owlto-dao/utils-go/alert"
)

type SwapTokenInfoManager struct {
	allTokens []*TokenInfo
	db        *sql.DB
	alerter   alert.Alerter
	mutex     *sync.RWMutex
}

func NewSwapTokenInfoManager(db *sql.DB, alerter alert.Alerter) *SwapTokenInfoManager {
	return &SwapTokenInfoManager{
		db:      db,
		alerter: alerter,
		mutex:   &sync.RWMutex{},
	}
}

func (mgr *SwapTokenInfoManager) GetByChainNameTokenAddr(chainName string, tokenAddr string) (*TokenInfo, bool) {
	var token TokenInfo
	err := mgr.db.QueryRow("SELECT token_name, chain_name, token_address, decimals, icon FROM t_swap_token_info where chain_name = ? and token_address = ?", chainName, tokenAddr).
		Scan(&token.TokenName, &token.ChainName, &token.TokenAddress, &token.Decimals, &token.Icon)
	if err != nil {
		return nil, false
	}
	return &token, true
}

func (mgr *SwapTokenInfoManager) GetByChainNameTokenName(chainName string, tokenName string) (*TokenInfo, bool) {
	var token TokenInfo
	err := mgr.db.QueryRow("SELECT token_name, chain_name, token_address, decimals, icon FROM t_swap_token_info where chain_name = ? and token_name = ?", chainName, tokenName).
		Scan(&token.TokenName, &token.ChainName, &token.TokenAddress, &token.Decimals, &token.Icon)
	if err != nil {
		return nil, false
	}
	return &token, true
}
