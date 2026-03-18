package loader

import (
	"database/sql"
	"math/big"
	"strings"
	"sync"

	"github.com/owlto-dao/utils-go/alert"
)

type TokenInfo struct {
	Id           int64
	TokenName    string
	ChainName    string
	ChainId      int64
	TokenAddress string
	Decimals     int32
	FullName     string
	TotalSupply  *big.Int
	Icon         string
	Url          string
}

type TokenInfoManager struct {
	chainNameTokenAddrs map[string]map[string]*TokenInfo
	chainNameTokenNames map[string]map[string]*TokenInfo
	chainIdTokenAddrs   map[int64]map[string]*TokenInfo
	chainIdTokenNames   map[int64]map[string]*TokenInfo
	allTokens           []*TokenInfo
	db                  *sql.DB
	alerter             alert.Alerter
	mutex               *sync.RWMutex
}

func NewTokenInfoManager(db *sql.DB, alerter alert.Alerter) *TokenInfoManager {
	return &TokenInfoManager{
		chainNameTokenAddrs: make(map[string]map[string]*TokenInfo),
		chainNameTokenNames: make(map[string]map[string]*TokenInfo),
		chainIdTokenAddrs:   make(map[int64]map[string]*TokenInfo),
		chainIdTokenNames:   make(map[int64]map[string]*TokenInfo),
		db:                  db,
		alerter:             alerter,
		mutex:               &sync.RWMutex{},
	}
}

func (mgr *TokenInfoManager) GetByChainNameTokenAddr(chainName string, tokenAddr string) (*TokenInfo, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	tokenAddrs, ok := mgr.chainNameTokenAddrs[strings.ToLower(strings.TrimSpace(chainName))]
	if ok {
		token, ok := tokenAddrs[strings.ToLower(strings.TrimSpace(tokenAddr))]
		return token, ok
	}
	return nil, false
}

func (mgr *TokenInfoManager) GetByChainIdTokenAddr(chainId int64, tokenAddr string) (*TokenInfo, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	tokenAddrs, ok := mgr.chainIdTokenAddrs[chainId]
	if ok {
		token, ok := tokenAddrs[strings.ToLower(strings.TrimSpace(tokenAddr))]
		return token, ok
	}
	return nil, false
}

func (mgr *TokenInfoManager) AddTokenInfo(token *TokenInfo) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	tokenAddrs, ok := mgr.chainNameTokenAddrs[strings.ToLower(token.ChainName)]
	if !ok {
		tokenAddrs = make(map[string]*TokenInfo)
		mgr.chainNameTokenAddrs[strings.ToLower(token.ChainName)] = tokenAddrs
	}
	tokenAddrs[strings.ToLower(token.TokenAddress)] = token

	tokenNames, ok := mgr.chainNameTokenNames[strings.ToLower(token.ChainName)]
	if !ok {
		tokenNames = make(map[string]*TokenInfo)
		mgr.chainNameTokenNames[strings.ToLower(token.ChainName)] = tokenNames
	}
	tokenNames[strings.ToLower(token.TokenName)] = token
}

func (mgr *TokenInfoManager) AddToken(chainName string, tokenName string, tokenAddr string, decimals int32) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	var token TokenInfo
	token.ChainName = strings.TrimSpace(chainName)
	token.TokenAddress = strings.TrimSpace(tokenAddr)
	token.TokenName = strings.TrimSpace(tokenName)
	token.Decimals = decimals

	tokenAddrs, ok := mgr.chainNameTokenAddrs[strings.ToLower(token.ChainName)]
	if !ok {
		tokenAddrs = make(map[string]*TokenInfo)
		mgr.chainNameTokenAddrs[strings.ToLower(token.ChainName)] = tokenAddrs
	}
	tokenAddrs[strings.ToLower(token.TokenAddress)] = &token

	tokenNames, ok := mgr.chainNameTokenNames[strings.ToLower(token.ChainName)]
	if !ok {
		tokenNames = make(map[string]*TokenInfo)
		mgr.chainNameTokenNames[strings.ToLower(token.ChainName)] = tokenNames
	}
	tokenNames[strings.ToLower(token.TokenName)] = &token
}

func (mgr *TokenInfoManager) GetByChainNameTokenName(chainName string, tokenName string) (*TokenInfo, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	tokenNames, ok := mgr.chainNameTokenNames[strings.ToLower(strings.TrimSpace(chainName))]
	if ok {
		token, ok := tokenNames[strings.ToLower(strings.TrimSpace(tokenName))]
		return token, ok
	}
	return nil, false
}

func (mgr *TokenInfoManager) GetByChainIdTokenName(chainId int64, tokenName string) (*TokenInfo, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	tokenNames, ok := mgr.chainIdTokenNames[chainId]
	if ok {
		token, ok := tokenNames[strings.ToLower(strings.TrimSpace(tokenName))]
		return token, ok
	}
	return nil, false
}

func (mgr *TokenInfoManager) GetTokenAddresses(chainName string) []string {
	addrs := make([]string, 0)
	mgr.mutex.RLock()
	tokenAddrs, ok := mgr.chainNameTokenAddrs[strings.ToLower(strings.TrimSpace(chainName))]
	if ok {
		for _, token := range tokenAddrs {
			addrs = append(addrs, token.TokenAddress)
		}
	}
	mgr.mutex.RUnlock()
	return addrs
}

func (mgr *TokenInfoManager) GetTokenAddressesByChainId(chainId int64) []string {
	addrs := make([]string, 0)
	mgr.mutex.RLock()
	tokenAddrs, ok := mgr.chainIdTokenAddrs[chainId]
	if ok {
		for _, token := range tokenAddrs {
			addrs = append(addrs, token.TokenAddress)
		}
	}
	mgr.mutex.RUnlock()
	return addrs
}

func (mgr *TokenInfoManager) GetAllTokens() []*TokenInfo {
	return mgr.allTokens
}

func (mgr *TokenInfoManager) LoadAllToken(chainManager *ChainInfoManager) {
	if chainManager == nil {
		panic("chainManager is required")
	}
	// Query the database to select only id and name fields
	rows, err := mgr.db.Query("SELECT id, token_name, chain_name, chain_id, token_address, decimals, icon FROM t_token_info")

	if err != nil || rows == nil {
		mgr.alerter.AlertText("select t_token_info error", err)
		return
	}

	defer rows.Close()

	chainNameTokenAddrs := make(map[string]map[string]*TokenInfo)
	chainNameTokenNames := make(map[string]map[string]*TokenInfo)
	chainIdTokenAddrs := make(map[int64]map[string]*TokenInfo)
	chainIdTokenNames := make(map[int64]map[string]*TokenInfo)
	allTokens := make([]*TokenInfo, 0)
	counter := 0

	// Iterate over the result set
	for rows.Next() {
		var token TokenInfo

		if err = rows.Scan(&token.Id, &token.TokenName, &token.ChainName, &token.ChainId, &token.TokenAddress, &token.Decimals, &token.Icon); err != nil {
			mgr.alerter.AlertText("scan t_token_info row error", err)
		} else {
			token.ChainName = strings.TrimSpace(token.ChainName)
			token.TokenAddress = strings.TrimSpace(token.TokenAddress)
			token.TokenName = strings.TrimSpace(token.TokenName)
			token.Icon = strings.TrimSpace(token.Icon)

			tokenAddrs, ok := chainNameTokenAddrs[strings.ToLower(token.ChainName)]
			if !ok {
				tokenAddrs = make(map[string]*TokenInfo)
				chainNameTokenAddrs[strings.ToLower(token.ChainName)] = tokenAddrs
			}
			tokenAddrs[strings.ToLower(token.TokenAddress)] = &token

			tokenNames, ok := chainNameTokenNames[strings.ToLower(token.ChainName)]
			if !ok {
				tokenNames = make(map[string]*TokenInfo)
				chainNameTokenNames[strings.ToLower(token.ChainName)] = tokenNames
			}
			tokenNames[strings.ToLower(token.TokenName)] = &token

			// Index by chainId
			if token.ChainId > 0 {
				tokenAddrsById, ok := chainIdTokenAddrs[token.ChainId]
				if !ok {
					tokenAddrsById = make(map[string]*TokenInfo)
					chainIdTokenAddrs[token.ChainId] = tokenAddrsById
				}
				tokenAddrsById[strings.ToLower(token.TokenAddress)] = &token

				tokenNamesById, ok := chainIdTokenNames[token.ChainId]
				if !ok {
					tokenNamesById = make(map[string]*TokenInfo)
					chainIdTokenNames[token.ChainId] = tokenNamesById
				}
				tokenNamesById[strings.ToLower(token.TokenName)] = &token
			}

			allTokens = append(allTokens, &token)
			counter++
		}
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		mgr.alerter.AlertText("get next t_token_info row error", err)
		return
	}

	allIDs := chainManager.GetChainInfoAutoIds()
	for _, id := range allIDs {
		chainInfo, ok := chainManager.GetChainInfoById(id)
		if !ok {
			continue
		}
		var token TokenInfo
		token.ChainName = chainInfo.Name
		token.TokenAddress = chainInfo.GasTokenAddress
		token.TokenName = chainInfo.GasTokenName
		token.Decimals = chainInfo.GasTokenDecimal
		token.Icon = chainInfo.GasTokenIcon

		tokenAddrs, ok := chainNameTokenAddrs[strings.ToLower(token.ChainName)]
		if !ok {
			tokenAddrs = make(map[string]*TokenInfo)
			chainNameTokenAddrs[strings.ToLower(token.ChainName)] = tokenAddrs
		}
		tokenNames, ok := chainNameTokenNames[strings.ToLower(token.ChainName)]
		if !ok {
			tokenNames = make(map[string]*TokenInfo)
			chainNameTokenNames[strings.ToLower(token.ChainName)] = tokenNames
		}
		_, ok = chainNameTokenNames[strings.ToLower(token.ChainName)][strings.ToLower(token.TokenName)]
		if !ok {
			tokenNames[strings.ToLower(token.TokenName)] = &token
			tokenAddrs[strings.ToLower(token.TokenAddress)] = &token
			allTokens = append(allTokens, &token)
		}
	}

	mgr.mutex.Lock()
	mgr.chainNameTokenAddrs = chainNameTokenAddrs
	mgr.chainNameTokenNames = chainNameTokenNames
	mgr.chainIdTokenAddrs = chainIdTokenAddrs
	mgr.chainIdTokenNames = chainIdTokenNames
	mgr.allTokens = allTokens
	mgr.mutex.Unlock()
}
