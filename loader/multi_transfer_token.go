package loader

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

type MultiTransferToken struct {
	Id           int64
	ChainName    string
	TokenName    string
	TokenAddress string
	Decimal      int32
	MaxValue     string
	Dtc          string
	Disabled     int8
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MultiTransferTokenManager struct {
	chainNameTokenAddrs map[string]map[string]*MultiTransferToken
	chainNameTokenNames map[string]map[string]*MultiTransferToken
	allTokens           []*MultiTransferToken

	db      *sql.DB
	alerter alert.Alerter
	mutex   *sync.RWMutex
}

func NewMultiTransferTokenManager(db *sql.DB, alerter alert.Alerter) *MultiTransferTokenManager {
	return &MultiTransferTokenManager{
		chainNameTokenAddrs: make(map[string]map[string]*MultiTransferToken),
		chainNameTokenNames: make(map[string]map[string]*MultiTransferToken),
		allTokens:           make([]*MultiTransferToken, 0),
		db:                  db,
		alerter:             alerter,
		mutex:               &sync.RWMutex{},
	}
}

func (mgr *MultiTransferTokenManager) LoadAllTokens() {
	rows, err := mgr.db.Query("SELECT id, chain_name, token_name, token_address, decimals, max_value, dtc, disabled, created_at, updated_at FROM t_multi_transfer_token")
	if err != nil {
		mgr.alerter.AlertText("select t_multi_transfer_token error", err)
		return
	}
	defer rows.Close()

	chainNameTokenAddrs := make(map[string]map[string]*MultiTransferToken)
	chainNameTokenNames := make(map[string]map[string]*MultiTransferToken)
	allTokens := make([]*MultiTransferToken, 0)

	for rows.Next() {
		var token MultiTransferToken
		if err := rows.Scan(&token.Id, &token.ChainName, &token.TokenName, &token.TokenAddress, &token.Decimal, &token.MaxValue, &token.Dtc, &token.Disabled, &token.CreatedAt, &token.UpdatedAt); err != nil {
			mgr.alerter.AlertText("scan t_multi_transfer_token row error", err)
			continue
		}

		token.ChainName = strings.TrimSpace(token.ChainName)
		token.TokenName = strings.TrimSpace(token.TokenName)
		token.TokenAddress = strings.TrimSpace(token.TokenAddress)
		token.MaxValue = strings.TrimSpace(token.MaxValue)
		token.Dtc = strings.TrimSpace(token.Dtc)

		// Group by ChainName -> TokenAddress
		tokenAddrs, ok := chainNameTokenAddrs[strings.ToLower(token.ChainName)]
		if !ok {
			tokenAddrs = make(map[string]*MultiTransferToken)
			chainNameTokenAddrs[strings.ToLower(token.ChainName)] = tokenAddrs
		}
		tokenAddrs[strings.ToLower(token.TokenAddress)] = &token

		// Group by ChainName -> TokenName
		tokenNames, ok := chainNameTokenNames[strings.ToLower(token.ChainName)]
		if !ok {
			tokenNames = make(map[string]*MultiTransferToken)
			chainNameTokenNames[strings.ToLower(token.ChainName)] = tokenNames
		}
		tokenNames[strings.ToLower(token.TokenName)] = &token

		allTokens = append(allTokens, &token)
	}

	if err := rows.Err(); err != nil {
		mgr.alerter.AlertText("iterate t_multi_transfer_token rows error", err)
		return
	}

	mgr.mutex.Lock()
	mgr.chainNameTokenAddrs = chainNameTokenAddrs
	mgr.chainNameTokenNames = chainNameTokenNames
	mgr.allTokens = allTokens
	mgr.mutex.Unlock()
}

func (mgr *MultiTransferTokenManager) GetByChainNameTokenAddr(chainName string, tokenAddr string) (*MultiTransferToken, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	tokenAddrs, ok := mgr.chainNameTokenAddrs[strings.ToLower(strings.TrimSpace(chainName))]
	if ok {
		token, ok := tokenAddrs[strings.ToLower(strings.TrimSpace(tokenAddr))]
		return token, ok
	}
	return nil, false
}

func (mgr *MultiTransferTokenManager) GetByChainNameTokenName(chainName string, tokenName string) (*MultiTransferToken, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	tokenNames, ok := mgr.chainNameTokenNames[strings.ToLower(strings.TrimSpace(chainName))]
	if ok {
		token, ok := tokenNames[strings.ToLower(strings.TrimSpace(tokenName))]
		return token, ok
	}
	return nil, false
}

func (mgr *MultiTransferTokenManager) GetAllTokens() []*MultiTransferToken {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	return mgr.allTokens
}
