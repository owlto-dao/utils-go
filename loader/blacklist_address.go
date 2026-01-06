package loader

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

type BlacklistAddress struct {
	Id        int64
	Address   string
	RiskDesc  string
	Status    int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BlacklistAddressManager struct {
	idBlacklists      map[int64]*BlacklistAddress
	addressBlacklists map[string]*BlacklistAddress
	db                *sql.DB
	alerter           alert.Alerter
	mutex             *sync.RWMutex
}

func NewBlacklistAddressManager(db *sql.DB, alerter alert.Alerter) *BlacklistAddressManager {
	return &BlacklistAddressManager{
		idBlacklists:      make(map[int64]*BlacklistAddress),
		addressBlacklists: make(map[string]*BlacklistAddress),
		db:                db,
		alerter:           alerter,
		mutex:             &sync.RWMutex{},
	}
}

func (mgr *BlacklistAddressManager) GetBlacklistById(id int64) (*BlacklistAddress, bool) {
	mgr.mutex.RLock()
	blacklist, ok := mgr.idBlacklists[id]
	mgr.mutex.RUnlock()
	return blacklist, ok
}

func (mgr *BlacklistAddressManager) IsBlacklisted(address string) bool {
	mgr.mutex.RLock()
	blacklist, ok := mgr.addressBlacklists[strings.ToLower(strings.TrimSpace(address))]
	mgr.mutex.RUnlock()
	return ok && blacklist.Status == 1
}

func (mgr *BlacklistAddressManager) GetBlacklistByAddress(address string) (*BlacklistAddress, bool) {
	mgr.mutex.RLock()
	blacklist, ok := mgr.addressBlacklists[strings.ToLower(strings.TrimSpace(address))]
	mgr.mutex.RUnlock()
	if ok && blacklist.Status == 1 {
		return blacklist, true
	}
	return nil, false
}

func (mgr *BlacklistAddressManager) GetAllBlacklists() []*BlacklistAddress {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	blacklists := make([]*BlacklistAddress, 0, len(mgr.addressBlacklists))
	for _, blacklist := range mgr.addressBlacklists {
		if blacklist.Status == 1 {
			blacklists = append(blacklists, blacklist)
		}
	}
	return blacklists
}

func (mgr *BlacklistAddressManager) LoadAllBlacklists() {
	// Query the database to select all fields
	rows, err := mgr.db.Query("SELECT id, address, risk_desc, status, created_at, updated_at FROM t_blacklist_address")

	if err != nil || rows == nil {
		mgr.alerter.AlertText("select t_blacklist_address error", err)
		return
	}

	defer rows.Close()

	idBlacklists := make(map[int64]*BlacklistAddress)
	addressBlacklists := make(map[string]*BlacklistAddress)

	// Iterate over the result set
	for rows.Next() {
		var blacklist BlacklistAddress
		if err := rows.Scan(&blacklist.Id, &blacklist.Address, &blacklist.RiskDesc, &blacklist.Status, &blacklist.CreatedAt, &blacklist.UpdatedAt); err != nil {
			mgr.alerter.AlertText("scan t_blacklist_address row error", err)
		} else {
			blacklist.Address = strings.TrimSpace(blacklist.Address)
			blacklist.RiskDesc = strings.TrimSpace(blacklist.RiskDesc)

			idBlacklists[blacklist.Id] = &blacklist
			addressBlacklists[strings.ToLower(blacklist.Address)] = &blacklist
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		mgr.alerter.AlertText("get next t_blacklist_address row error", err)
		return
	}

	mgr.mutex.Lock()
	mgr.idBlacklists = idBlacklists
	mgr.addressBlacklists = addressBlacklists
	mgr.mutex.Unlock()
}
