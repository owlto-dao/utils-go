package loader

import (
	"database/sql"
	"sort"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

type NodeInfo struct {
	Id              int64
	UpdateTimestamp time.Time
	InsertTimestamp time.Time
	ChainId         int64
	RpcURL          string
	Type            int32
	Usability       int32
}

type NodeInfoManager struct {
	idNodes          map[int64]*NodeInfo
	chainIdNodes     map[int64][]*NodeInfo
	chainIdTypeNodes map[int64]map[int32][]*NodeInfo
	allNodes         []*NodeInfo

	db      *sql.DB
	alerter alert.Alerter
	mutex   *sync.RWMutex
}

func NewNodeInfoManager(db *sql.DB, alerter alert.Alerter) *NodeInfoManager {
	return &NodeInfoManager{
		idNodes:          make(map[int64]*NodeInfo),
		chainIdNodes:     make(map[int64][]*NodeInfo),
		chainIdTypeNodes: make(map[int64]map[int32][]*NodeInfo),
		allNodes:         make([]*NodeInfo, 0, 64),
		db:               db,
		alerter:          alerter,
		mutex:            &sync.RWMutex{},
	}
}

func (mgr *NodeInfoManager) GetNodeInfoById(id int64) (*NodeInfo, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	node, ok := mgr.idNodes[id]
	return node, ok
}

func (mgr *NodeInfoManager) GetNodeInfosByChainId(chainId int64) []*NodeInfo {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	nodes := mgr.chainIdNodes[chainId]
	out := make([]*NodeInfo, len(nodes))
	copy(out, nodes)
	return out
}

func (mgr *NodeInfoManager) GetNodeInfosByChainIdAndType(chainId int64, nodeType int32) []*NodeInfo {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	typeNodes, ok := mgr.chainIdTypeNodes[chainId]
	if !ok {
		return nil
	}
	nodes := typeNodes[nodeType]
	out := make([]*NodeInfo, len(nodes))
	copy(out, nodes)
	return out
}

func (mgr *NodeInfoManager) GetAllNodes() []*NodeInfo {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	nodes := make([]*NodeInfo, len(mgr.allNodes))
	copy(nodes, mgr.allNodes)
	return nodes
}

func (mgr *NodeInfoManager) GetEnabledNodes() []*NodeInfo {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	enabled := make([]*NodeInfo, 0, len(mgr.allNodes))
	for _, node := range mgr.allNodes {
		if node.Usability > 0 {
			enabled = append(enabled, node)
		}
	}
	return enabled
}

// GetBestNodeByChainIdAndType 获取指定链和节点类型下可用性最高的节点。
func (mgr *NodeInfoManager) GetBestNodeByChainIdAndType(chainId int64, nodeType int32) (*NodeInfo, bool) {
	nodes := mgr.GetNodeInfosByChainIdAndType(chainId, nodeType)
	if len(nodes) == 0 {
		return nil, false
	}
	return nodes[0], true
}

// GetAvailableNodeByChainId 根据 chainId 获取一个可用节点（usability > 0，按可用性降序优先）。
func (mgr *NodeInfoManager) GetAvailableNodeByChainId(chainId int64) (*NodeInfo, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	nodes := mgr.chainIdNodes[chainId]
	for _, node := range nodes {
		if node.Usability > 0 {
			return node, true
		}
	}
	return nil, false
}

func (mgr *NodeInfoManager) LoadAllNodes() {
	rows, err := mgr.db.Query(`
		SELECT id, update_timestamp, insert_timestamp, chain_id, rpc_url, type, usability
		FROM t_node_info
	`)
	if err != nil || rows == nil {
		mgr.alerter.AlertText("select t_node_info error", err)
		return
	}
	defer rows.Close()

	idNodes := make(map[int64]*NodeInfo)
	chainIdNodes := make(map[int64][]*NodeInfo)
	chainIdTypeNodes := make(map[int64]map[int32][]*NodeInfo)
	allNodes := make([]*NodeInfo, 0, 64)

	for rows.Next() {
		var node NodeInfo
		if err := rows.Scan(
			&node.Id,
			&node.UpdateTimestamp,
			&node.InsertTimestamp,
			&node.ChainId,
			&node.RpcURL,
			&node.Type,
			&node.Usability,
		); err != nil {
			mgr.alerter.AlertText("scan t_node_info row error", err)
			continue
		}

		idNodes[node.Id] = &node
		chainIdNodes[node.ChainId] = append(chainIdNodes[node.ChainId], &node)
		if _, ok := chainIdTypeNodes[node.ChainId]; !ok {
			chainIdTypeNodes[node.ChainId] = make(map[int32][]*NodeInfo)
		}
		chainIdTypeNodes[node.ChainId][node.Type] = append(chainIdTypeNodes[node.ChainId][node.Type], &node)
		allNodes = append(allNodes, &node)
	}

	if err := rows.Err(); err != nil {
		mgr.alerter.AlertText("iterate t_node_info rows error", err)
		return
	}

	sortNodesByUsability(allNodes)
	for chainId := range chainIdNodes {
		sortNodesByUsability(chainIdNodes[chainId])
	}
	for chainId := range chainIdTypeNodes {
		for nodeType := range chainIdTypeNodes[chainId] {
			sortNodesByUsability(chainIdTypeNodes[chainId][nodeType])
		}
	}

	mgr.mutex.Lock()
	mgr.idNodes = idNodes
	mgr.chainIdNodes = chainIdNodes
	mgr.chainIdTypeNodes = chainIdTypeNodes
	mgr.allNodes = allNodes
	mgr.mutex.Unlock()
}

func sortNodesByUsability(nodes []*NodeInfo) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Usability == nodes[j].Usability {
			return nodes[i].Id < nodes[j].Id
		}
		return nodes[i].Usability > nodes[j].Usability
	})
}
