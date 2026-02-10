package loader

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

type MultiTransferChain struct {
	Id        int64
	ChainId   int32
	ChainName string
	Disabled  int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MultiTransferChainManager struct {
	idChains      map[int64]*MultiTransferChain
	chainIdChains map[int32]*MultiTransferChain
	allChains     []*MultiTransferChain

	db      *sql.DB
	alerter alert.Alerter
	mutex   *sync.RWMutex
}

func NewMultiTransferChainManager(db *sql.DB, alerter alert.Alerter) *MultiTransferChainManager {
	return &MultiTransferChainManager{
		idChains:      make(map[int64]*MultiTransferChain),
		chainIdChains: make(map[int32]*MultiTransferChain),
		allChains:     make([]*MultiTransferChain, 0),
		db:            db,
		alerter:       alerter,
		mutex:         &sync.RWMutex{},
	}
}

func (mgr *MultiTransferChainManager) LoadAllChains() {
	rows, err := mgr.db.Query("SELECT id, chain_id, chain_name, disabled, created_at, updated_at FROM t_multi_transfer_chain")
	if err != nil {
		mgr.alerter.AlertText("select t_multi_transfer_chain error", err)
		return
	}
	defer rows.Close()

	idChains := make(map[int64]*MultiTransferChain)
	chainIdChains := make(map[int32]*MultiTransferChain)
	allChains := make([]*MultiTransferChain, 0)

	for rows.Next() {
		var chain MultiTransferChain
		if err := rows.Scan(&chain.Id, &chain.ChainId, &chain.ChainName, &chain.Disabled, &chain.CreatedAt, &chain.UpdatedAt); err != nil {
			mgr.alerter.AlertText("scan t_multi_transfer_chain row error", err)
			continue
		}

		chain.ChainName = strings.TrimSpace(chain.ChainName)

		idChains[chain.Id] = &chain
		chainIdChains[chain.ChainId] = &chain
		allChains = append(allChains, &chain)
	}

	if err := rows.Err(); err != nil {
		mgr.alerter.AlertText("iterate t_multi_transfer_chain rows error", err)
		return
	}

	mgr.mutex.Lock()
	mgr.idChains = idChains
	mgr.chainIdChains = chainIdChains
	mgr.allChains = allChains
	mgr.mutex.Unlock()
}

func (mgr *MultiTransferChainManager) GetChainById(id int64) (*MultiTransferChain, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	chain, ok := mgr.idChains[id]
	return chain, ok
}

func (mgr *MultiTransferChainManager) GetChainByChainId(chainId int32) (*MultiTransferChain, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	chain, ok := mgr.chainIdChains[chainId]
	return chain, ok
}

func (mgr *MultiTransferChainManager) GetAllChains() []*MultiTransferChain {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	return mgr.allChains
}
