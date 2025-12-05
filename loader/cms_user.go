package loader

import (
	"database/sql"
	"github.com/owlto-dao/utils-go/log"
	"strings"
	"sync"
)

type DevRole struct {
	UserId int32 `json:"user_id"`
	RoleId int32 `json:"role_id"`
}

type CmsRole struct {
	Id   int64
	Name string
}

type CmsUser struct {
	Id      int64
	Name    string
	Address string
	Roles   []*CmsRole
}

type CmsUserManager struct {
	allUsers []*CmsUser
	db       *sql.DB
	mutex    *sync.RWMutex
}

func NewCmsUserManager(db *sql.DB) *CmsUserManager {
	return &CmsUserManager{
		allUsers: []*CmsUser{},
		db:       db,
		mutex:    &sync.RWMutex{},
	}
}

func (mgr *CmsUserManager) HasRole(userAddress string, roleName string) bool {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	for _, user := range mgr.allUsers {
		if strings.ToLower(user.Address) == strings.ToLower(userAddress) {
			for _, role := range user.Roles {
				if role.Name == roleName {
					return true
				}
			}
		}
	}
	return false
}

func (mgr *CmsUserManager) LoadAllCmsUsers() {
	allUsers := []*CmsUser{}
	allRoles := []*CmsRole{}

	rows, err := mgr.db.Query("SELECT id, name FROM dev_role")
	if err != nil || rows == nil {
		log.Error("query dev role failed ", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var role CmsRole
		if err = rows.Scan(&role.Id, &role.Name); err != nil {
			log.Error("scan dev_role row error", err)
			return
		} else {
			allRoles = append(allRoles, &role)
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("get next dev_role row error", err)
		return
	}

	adminRows, err := mgr.db.Query("SELECT id, name, address FROM dev_white_admin")
	if err != nil || adminRows == nil {
		log.Error("query dev_white_admin failed ", err)
		return
	}
	defer adminRows.Close()
	for adminRows.Next() {
		var user CmsUser
		if err = adminRows.Scan(&user.Id, &user.Name, &user.Address); err != nil {
			log.Error("scan dev_white_admin failed ", err)
			return
		} else {
			user.Roles = []*CmsRole{}
			mgr.allUsers = append(mgr.allUsers, &user)
		}
	}
	if err = adminRows.Err(); err != nil {
		log.Error("get next dev_white_admin failed ", err)
		return
	}

	devRoleRows, err := mgr.db.Query("SELECT user_id, role_id FROM dev_roles")
	if err != nil || devRoleRows == nil {
		log.Error("query dev_roles failed ", err)
		return
	}
	defer devRoleRows.Close()
	for devRoleRows.Next() {
		var devRole DevRole
		if err = devRoleRows.Scan(&devRole.UserId, &devRole.RoleId); err != nil {
			log.Error("scan dev_role row error", err)
			return
		} else {
			for _, user := range allUsers {
				for _, role := range allRoles {
					if user.Id == int64(devRole.UserId) && role.Id == int64(devRole.RoleId) {
						user.Roles = append(user.Roles, role)
					}
				}
			}
		}
	}
	if err = devRoleRows.Err(); err != nil {
		log.Error("get next dev_roles failed ", err)
		return
	}

	mgr.mutex.Lock()
	mgr.allUsers = allUsers
	mgr.mutex.Unlock()
}
