package loader

import (
	"database/sql"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

type BlacklistChain struct {
	Id        int64
	ChainId   int64
	Status    int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BlacklistChainManager struct {
	idBlacklists      map[int64]*BlacklistChain
	chainIdBlacklists map[int64]*BlacklistChain
	db                *sql.DB
	alerter           alert.Alerter
	mutex             *sync.RWMutex
}

func NewBlacklistChainManager(db *sql.DB, alerter alert.Alerter) *BlacklistChainManager {
	return &BlacklistChainManager{
		idBlacklists:      make(map[int64]*BlacklistChain),
		chainIdBlacklists: make(map[int64]*BlacklistChain),
		db:                db,
		alerter:           alerter,
		mutex:             &sync.RWMutex{},
	}
}

func (mgr *BlacklistChainManager) GetBlacklistById(id int64) (*BlacklistChain, bool) {
	mgr.mutex.RLock()
	blacklist, ok := mgr.idBlacklists[id]
	mgr.mutex.RUnlock()
	return blacklist, ok
}

func (mgr *BlacklistChainManager) IsBlacklisted(chainId int64) bool {
	mgr.mutex.RLock()
	blacklist, ok := mgr.chainIdBlacklists[chainId]
	mgr.mutex.RUnlock()
	return ok && blacklist.Status == 1
}

func (mgr *BlacklistChainManager) GetBlacklistByChainId(chainId int64) (*BlacklistChain, bool) {
	mgr.mutex.RLock()
	blacklist, ok := mgr.chainIdBlacklists[chainId]
	mgr.mutex.RUnlock()
	if ok && blacklist.Status == 1 {
		return blacklist, true
	}
	return nil, false
}

func (mgr *BlacklistChainManager) GetAllBlacklists() []*BlacklistChain {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	blacklists := make([]*BlacklistChain, 0, len(mgr.chainIdBlacklists))
	for _, blacklist := range mgr.chainIdBlacklists {
		if blacklist.Status == 1 {
			blacklists = append(blacklists, blacklist)
		}
	}
	return blacklists
}

func (mgr *BlacklistChainManager) LoadAllBlacklists() {
	// Query the database to select all fields
	rows, err := mgr.db.Query("SELECT id, chain_id, status, created_at, updated_at FROM t_blacklist_chain")

	if err != nil || rows == nil {
		mgr.alerter.AlertText("select t_blacklist_chain error", err)
		return
	}

	defer rows.Close()

	idBlacklists := make(map[int64]*BlacklistChain)
	chainIdBlacklists := make(map[int64]*BlacklistChain)

	// Iterate over the result set
	for rows.Next() {
		var blacklist BlacklistChain
		if err := rows.Scan(&blacklist.Id, &blacklist.ChainId, &blacklist.Status, &blacklist.CreatedAt, &blacklist.UpdatedAt); err != nil {
			mgr.alerter.AlertText("scan t_blacklist_chain row error", err)
		} else {
			idBlacklists[blacklist.Id] = &blacklist
			chainIdBlacklists[blacklist.ChainId] = &blacklist
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		mgr.alerter.AlertText("get next t_blacklist_chain row error", err)
		return
	}

	mgr.mutex.Lock()
	mgr.idBlacklists = idBlacklists
	mgr.chainIdBlacklists = chainIdBlacklists
	mgr.mutex.Unlock()
}
