package loader

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

// AggregatorID 聚合器标识
type AggregatorID int

const (
	AggregatorNone     AggregatorID = 0
	AggregatorAcross   AggregatorID = 1
	AggregatorRelay    AggregatorID = 2
	AggregatorDebridge AggregatorID = 3
)

func (a AggregatorID) String() string {
	switch a {
	case AggregatorAcross:
		return "Across"
	case AggregatorRelay:
		return "Relay"
	case AggregatorDebridge:
		return "Debridge"
	default:
		return "Unknown"
	}
}

// AggregatorConfig 聚合器基础配置（对应 t_aggregate_config）
type AggregatorConfig struct {
	ID         AggregatorID
	Name       string
	IsEnabled  bool
	Priority   int
	APIBaseURL string
}

// AggregateChainConfig 聚合器链路配置（对应 t_aggregate_chain_config）
type AggregateChainConfig struct {
	ID                     int64
	AggregateID            AggregatorID
	ChainName              string
	ChainID                int64
	DepositContractAddress string
	BridgeFeeRateBps       int
	FillDeadlineSeconds    int
	SwapperAddress         string
	IsEnabled              bool
	Priority               int
}

// RouteConfig 路由配置（对应 t_aggregator_route_config）
type RouteConfig struct {
	ID               int64
	AggregateID      AggregatorID
	FromChainID      int64
	FromChainName    string
	FromTokenAddress string
	FromTokenSymbol  string
	ToChainID        int64
	ToChainName      string
	ToTokenAddress   string
	ToTokenSymbol    string
	IsNative         bool
	MinAmount        string // UI 金额
	MaxAmount        string // UI 金额，"0" 表示无上限
	IsEnabled        bool
	Priority         int
}

// FeeSegment 分段收费配置（对应 t_aggregator_route_fee_segment）
type FeeSegment struct {
	ID                 int64
	RouteID            AggregatorID // 关联 RouteConfig.ID
	MinAmountUI        string       // UI 金额下限
	MaxAmountUI        string       // UI 金额上限，"0" 表示无上限
	OwltoFeeFixedUI    string       // Owlto 固定加收（UI 金额）
	OwltoFeeRateBps    int          // Owlto 百分比加收（基点，10=0.1%）
	ProtocolFeeRateBps int          // 协议官方费率（基点）
	IsEnabled          bool
	Priority           int
}

// FeeResult 费用计算结果
type FeeResult struct {
	BestAggregator     AggregatorID // 最优聚合器
	ChannelFee         *big.Int     // 协议费用（最小单位）
	OwltoFee           *big.Int     // Owlto 加收费用（最小单位）
	TotalGasFee        *big.Int     // 总 Gas Fee = ChannelFee + OwltoFee
	ProtocolFee        *big.Int     // 协议官方费率（如 Across 0.1%）
	ChannelFeeUI       string       // 协议费用（UI 金额）
	OwltoFeeUI         string       // Owlto 加收费用（UI 金额）
	TotalGasFeeUI      string       // 总费用（UI 金额）
	ProtocolFeeRateBps int          // 协议费率（基点）
}

// RouteQueryResult 路由查询结果
type RouteQueryResult struct {
	Route       *RouteConfig
	FeeSegments []*FeeSegment
	IsAvailable bool
}

// BestRouteResult 最优路由选择结果
type BestRouteResult struct {
	Aggregator   AggregatorID
	Route        *RouteConfig
	FeeSegment   *FeeSegment
	FeeResult    *FeeResult
	IsAggregator bool // 是否走聚合器（非 Owlto Native）
}

// AggregatorManager 聚合器统一管理器
type AggregatorManager struct {
	// 聚合器配置
	aggregatorConfigs  map[AggregatorID]*AggregatorConfig
	chainConfigs       map[AggregatorID]map[int64]*AggregateChainConfig  // aggregateID -> chainID -> config
	chainConfigsByName map[AggregatorID]map[string]*AggregateChainConfig // aggregateID -> chainName(lower) -> config

	// 路由索引（多维）
	routesByChainPair map[int64]map[int64][]*RouteConfig // fromChainID -> toChainID -> routes
	routesBySymbol    map[string][]*RouteConfig          // tokenSymbol(lower) -> routes
	routesByID        map[int64]*RouteConfig             // routeID -> route

	// 分段收费索引
	feeSegmentsByRouteID map[AggregatorID][]*FeeSegment // routeID -> segments (已按 minAmount 排序)

	// 基础设施
	db      *sql.DB
	alerter alert.Alerter
	mutex   *sync.RWMutex

	// 定时刷新
	refreshInterval time.Duration
	stopCh          chan struct{}
}

func NewAggregatorManager(db *sql.DB, alerter alert.Alerter) *AggregatorManager {
	return &AggregatorManager{
		aggregatorConfigs:    make(map[AggregatorID]*AggregatorConfig),
		chainConfigs:         make(map[AggregatorID]map[int64]*AggregateChainConfig),
		chainConfigsByName:   make(map[AggregatorID]map[string]*AggregateChainConfig),
		routesByChainPair:    make(map[int64]map[int64][]*RouteConfig),
		routesBySymbol:       make(map[string][]*RouteConfig),
		routesByID:           make(map[int64]*RouteConfig),
		feeSegmentsByRouteID: make(map[AggregatorID][]*FeeSegment),
		db:                   db,
		alerter:              alerter,
		mutex:                new(sync.RWMutex),
		refreshInterval:      5 * time.Minute,
	}
}

// LoadAll 加载所有配置到内存
func (mgr *AggregatorManager) LoadAll() error {
	if err := mgr.loadAggregatorConfigs(); err != nil {
		return fmt.Errorf("load aggregator configs: %w", err)
	}
	if err := mgr.loadAggregateChainConfigs(); err != nil {
		return fmt.Errorf("load aggregate chain configs: %w", err)
	}
	if err := mgr.loadRouteConfigs(); err != nil {
		return fmt.Errorf("load route configs: %w", err)
	}
	if err := mgr.loadFeeSegments(); err != nil {
		return fmt.Errorf("load fee segments: %w", err)
	}
	return nil
}

// loadAggregatorConfigs 从 t_aggregate_config 加载聚合器配置
func (mgr *AggregatorManager) loadAggregatorConfigs() error {
	rows, err := mgr.db.Query(`
        SELECT id, name, is_enabled, priority, api_base_url
        FROM t_aggregate_config
        WHERE is_enabled = 1
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	newConfigs := make(map[AggregatorID]*AggregatorConfig)
	for rows.Next() {
		var cfg AggregatorConfig
		var id int
		if err := rows.Scan(&id, &cfg.Name, &cfg.IsEnabled, &cfg.Priority, &cfg.APIBaseURL); err != nil {
			mgr.alerter.AlertText("scan aggregator config error", err)
			continue
		}
		cfg.ID = AggregatorID(id)
		newConfigs[cfg.ID] = &cfg
	}

	mgr.mutex.Lock()
	mgr.aggregatorConfigs = newConfigs
	mgr.mutex.Unlock()
	return nil
}

// loadAggregateChainConfigs 从 t_aggregate_chain_config 加载聚合器链路配置
func (mgr *AggregatorManager) loadAggregateChainConfigs() error {
	rows, err := mgr.db.Query(`
        SELECT id, aggregate_id, chain_name, chain_id, deposit_contract_address,
               bridge_fee_rate_bps, fill_deadline_seconds, swapper_address,
               is_enabled, priority
        FROM t_aggregate_chain_config
        WHERE is_enabled = 1
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	newByAggID := make(map[AggregatorID]map[int64]*AggregateChainConfig)
	newByAggName := make(map[AggregatorID]map[string]*AggregateChainConfig)

	for rows.Next() {
		var cfg AggregateChainConfig
		var aggID int
		var depositContractAddress sql.NullString
		var bridgeFeeRateBps sql.NullInt64
		var fillDeadlineSeconds sql.NullInt64
		var swapperAddress sql.NullString

		if err := rows.Scan(
			&cfg.ID, &aggID, &cfg.ChainName, &cfg.ChainID, &depositContractAddress,
			&bridgeFeeRateBps, &fillDeadlineSeconds, &swapperAddress,
			&cfg.IsEnabled, &cfg.Priority,
		); err != nil {
			mgr.alerter.AlertText("scan aggregate chain config error", err)
			continue
		}

		cfg.AggregateID = AggregatorID(aggID)
		if depositContractAddress.Valid {
			cfg.DepositContractAddress = depositContractAddress.String
		}
		if bridgeFeeRateBps.Valid {
			cfg.BridgeFeeRateBps = int(bridgeFeeRateBps.Int64)
		}
		if fillDeadlineSeconds.Valid {
			cfg.FillDeadlineSeconds = int(fillDeadlineSeconds.Int64)
		}
		if swapperAddress.Valid {
			cfg.SwapperAddress = swapperAddress.String
		}

		if _, ok := newByAggID[cfg.AggregateID]; !ok {
			newByAggID[cfg.AggregateID] = make(map[int64]*AggregateChainConfig)
		}
		newByAggID[cfg.AggregateID][cfg.ChainID] = &cfg

		if _, ok := newByAggName[cfg.AggregateID]; !ok {
			newByAggName[cfg.AggregateID] = make(map[string]*AggregateChainConfig)
		}
		chainNameKey := strings.ToLower(strings.TrimSpace(cfg.ChainName))
		newByAggName[cfg.AggregateID][chainNameKey] = &cfg
	}

	mgr.mutex.Lock()
	mgr.chainConfigs = newByAggID
	mgr.chainConfigsByName = newByAggName
	mgr.mutex.Unlock()
	return nil
}

// loadRouteConfigs 从 t_aggregator_route_config 加载路由
func (mgr *AggregatorManager) loadRouteConfigs() error {
	rows, err := mgr.db.Query(`
        SELECT id, aggregate_id, from_chain_id, from_chain_name, from_token_address, from_token_symbol,
               to_chain_id, to_chain_name, to_token_address, to_token_symbol,
               is_native, min_amount, max_amount, is_enabled, priority
        FROM t_aggregate_route_config
        WHERE is_enabled = 1
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	newByChainPair := make(map[int64]map[int64][]*RouteConfig)
	newBySymbol := make(map[string][]*RouteConfig)
	newByID := make(map[int64]*RouteConfig)

	for rows.Next() {
		var r RouteConfig
		var aggID int
		if err := rows.Scan(
			&r.ID, &aggID, &r.FromChainID, &r.FromChainName, &r.FromTokenAddress, &r.FromTokenSymbol,
			&r.ToChainID, &r.ToChainName, &r.ToTokenAddress, &r.ToTokenSymbol,
			&r.IsNative, &r.MinAmount, &r.MaxAmount, &r.IsEnabled, &r.Priority,
		); err != nil {
			mgr.alerter.AlertText("scan route config error", err)
			continue
		}
		r.AggregateID = AggregatorID(aggID)

		// 索引：chainPair
		if _, ok := newByChainPair[r.FromChainID]; !ok {
			newByChainPair[r.FromChainID] = make(map[int64][]*RouteConfig)
		}
		newByChainPair[r.FromChainID][r.ToChainID] = append(newByChainPair[r.FromChainID][r.ToChainID], &r)

		// 索引：symbol
		sym := strings.ToLower(r.FromTokenSymbol)
		newBySymbol[sym] = append(newBySymbol[sym], &r)

		// 索引：ID
		newByID[r.ID] = &r
	}

	mgr.mutex.Lock()
	mgr.routesByChainPair = newByChainPair
	mgr.routesBySymbol = newBySymbol
	mgr.routesByID = newByID
	mgr.mutex.Unlock()
	return nil
}

// loadFeeSegments 从 t_aggregator_route_fee_segment 加载分段收费
func (mgr *AggregatorManager) loadFeeSegments() error {
	rows, err := mgr.db.Query(`
        SELECT id, route_id, min_amount_ui, max_amount_ui,
               owlto_fee_fixed_ui, owlto_fee_rate_bps, protocol_fee_rate_bps,
               is_enabled, priority
        FROM t_aggregate_route_fee_segment
        WHERE is_enabled = 1
        ORDER BY route_id, CAST(min_amount_ui AS DECIMAL(65,18)) ASC
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	newByRouteID := make(map[AggregatorID][]*FeeSegment)
	for rows.Next() {
		var seg FeeSegment
		if err := rows.Scan(
			&seg.ID, &seg.RouteID, &seg.MinAmountUI, &seg.MaxAmountUI,
			&seg.OwltoFeeFixedUI, &seg.OwltoFeeRateBps, &seg.ProtocolFeeRateBps,
			&seg.IsEnabled, &seg.Priority,
		); err != nil {
			mgr.alerter.AlertText("scan fee segment error", err)
			continue
		}
		newByRouteID[seg.RouteID] = append(newByRouteID[seg.RouteID], &seg)
	}

	mgr.mutex.Lock()
	mgr.feeSegmentsByRouteID = newByRouteID
	mgr.mutex.Unlock()
	return nil
}

// GetAggregateChainConfigByChainID 根据 aggregateID + chainID 查询链路配置
func (mgr *AggregatorManager) GetAggregateChainConfigByChainID(aggregateID AggregatorID, chainID int64) *AggregateChainConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if chainMap, ok := mgr.chainConfigs[aggregateID]; ok {
		return chainMap[chainID]
	}
	return nil
}

// GetAggregateChainConfigByChainName 根据 aggregateID + chainName 查询链路配置
func (mgr *AggregatorManager) GetAggregateChainConfigByChainName(aggregateID AggregatorID, chainName string) *AggregateChainConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if chainMap, ok := mgr.chainConfigsByName[aggregateID]; ok {
		return chainMap[strings.ToLower(strings.TrimSpace(chainName))]
	}
	return nil
}

// GetRoutesByChainPair 根据链对查询所有可用路由
func (mgr *AggregatorManager) GetRoutesByChainPair(fromChainID, toChainID int64) []*RouteConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if destMap, ok := mgr.routesByChainPair[fromChainID]; ok {
		if routes, ok := destMap[toChainID]; ok {
			return routes
		}
	}
	return nil
}

// GetRoutesByChainPairAndSymbol 根据链对+Token符号查询路由
func (mgr *AggregatorManager) GetRoutesByChainPairAndSymbol(
	fromChainID, toChainID int64,
	tokenSymbol string,
) []*RouteConfig {
	routes := mgr.GetRoutesByChainPair(fromChainID, toChainID)
	if routes == nil {
		return nil
	}

	sym := strings.ToLower(strings.TrimSpace(tokenSymbol))
	var result []*RouteConfig
	for _, r := range routes {
		if strings.ToLower(r.FromTokenSymbol) == sym || strings.ToLower(r.ToTokenSymbol) == sym {
			result = append(result, r)
		}
	}
	return result
}

// GetFeeSegments 获取路由的分段收费配置
func (mgr *AggregatorManager) GetFeeSegments(routeID AggregatorID) []*FeeSegment {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	return mgr.feeSegmentsByRouteID[routeID]
}

// FindBestRoute 根据金额选择最优路由
// amountUI: UI 金额（如 "12.5" USDT）
// tokenSymbol: Token 符号（如 "USDT"、"ETH"）
// fromChainID, toChainID: 链 ID
// decimals: Token 精度，用于将 UI 金额转为最小单位
func (mgr *AggregatorManager) FindBestRoute(
	fromChainID, toChainID int64,
	tokenSymbol string,
	amountUI string,
	decimals int32,
) (*BestRouteResult, error) {
	// 1. 查询该链对+Token 的所有路由
	routes := mgr.GetRoutesByChainPairAndSymbol(fromChainID, toChainID, tokenSymbol)
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found for %d->%d %s", fromChainID, toChainID, tokenSymbol)
	}

	// 2. 解析 UI 金额
	amtRat, ok := new(big.Rat).SetString(amountUI)
	if !ok || amtRat.Sign() <= 0 {
		return nil, fmt.Errorf("invalid amount: %s", amountUI)
	}

	// 3. 遍历路由，查找匹配的 FeeSegment
	var bestResult *BestRouteResult
	var lowestFee *big.Rat

	for _, route := range routes {
		// 检查金额是否在路由的 min/max 范围内
		if !mgr.isAmountInRouteRange(amtRat, route) {
			continue
		}

		// 查找该路由下匹配金额的 FeeSegment
		segment := mgr.findMatchingSegment(route.AggregateID, amtRat)
		if segment == nil {
			continue // 无匹配的分段配置
		}

		// 计算该分段的费用
		feeResult := mgr.calculateFee(amtRat, segment, decimals)
		if feeResult == nil {
			continue
		}

		// 比较找出最低费用
		feeRat := new(big.Rat).SetInt(feeResult.TotalGasFee)
		if lowestFee == nil || feeRat.Cmp(lowestFee) < 0 {
			lowestFee = feeRat
			bestResult = &BestRouteResult{
				Aggregator:   segment.RouteID,
				Route:        route,
				FeeSegment:   segment,
				FeeResult:    feeResult,
				IsAggregator: true,
			}
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("no matching fee segment for amount %s", amountUI)
	}

	return bestResult, nil
}

// isAmountInRouteRange 检查金额是否在路由的 min/max 范围内
func (mgr *AggregatorManager) isAmountInRouteRange(amtRat *big.Rat, route *RouteConfig) bool {
	// 解析 min
	minRat := new(big.Rat)
	if route.MinAmount != "" && route.MinAmount != "0" {
		if _, ok := minRat.SetString(route.MinAmount); !ok {
			return false
		}
		if amtRat.Cmp(minRat) < 0 {
			return false
		}
	}

	// 解析 max（"0" 或空表示无上限）
	if route.MaxAmount != "" && route.MaxAmount != "0" {
		maxRat := new(big.Rat)
		if _, ok := maxRat.SetString(route.MaxAmount); !ok {
			return false
		}
		if amtRat.Cmp(maxRat) > 0 {
			return false
		}
	}

	return true
}

// findMatchingSegment 在分段列表中找到匹配金额的 segment
func (mgr *AggregatorManager) findMatchingSegment(routeID AggregatorID, amtRat *big.Rat) *FeeSegment {
	segments := mgr.GetFeeSegments(routeID)
	if segments == nil {
		return nil
	}

	for _, seg := range segments {
		// 解析 min
		minRat := new(big.Rat)
		if seg.MinAmountUI != "" {
			if _, ok := minRat.SetString(seg.MinAmountUI); !ok {
				continue
			}
		}

		// 解析 max
		maxRat := new(big.Rat)
		unlimited := seg.MaxAmountUI == "" || seg.MaxAmountUI == "0"
		if !unlimited {
			if _, ok := maxRat.SetString(seg.MaxAmountUI); !ok {
				continue
			}
		}

		// 判断 amt 是否在 [min, max) 区间内
		if amtRat.Cmp(minRat) >= 0 {
			if unlimited || amtRat.Cmp(maxRat) < 0 {
				return seg
			}
		}
	}

	return nil
}

// calculateFee 根据 FeeSegment 计算费用
// amtRat: UI 金额（big.Rat）
// seg: 匹配的分段配置
// decimals: Token 精度
func (mgr *AggregatorManager) calculateFee(amtRat *big.Rat, seg *FeeSegment, decimals int32) *FeeResult {
	result := &FeeResult{
		BestAggregator:     seg.RouteID,
		ProtocolFeeRateBps: seg.ProtocolFeeRateBps,
	}

	// 计算精度因子：10^decimals
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	decimalFactorRat := new(big.Rat).SetInt(decimalFactor)

	// 将 UI 金额转为最小单位
	amtWeiRat := new(big.Rat).Mul(amtRat, decimalFactorRat)
	amtWei := new(big.Int).Div(amtWeiRat.Num(), amtWeiRat.Denom())

	// 1. 计算 Owlto 固定费用
	owltoFixedUI := new(big.Rat)
	if seg.OwltoFeeFixedUI != "" && seg.OwltoFeeFixedUI != "0" {
		owltoFixedUI.SetString(seg.OwltoFeeFixedUI)
	}
	owltoFixedWeiRat := new(big.Rat).Mul(owltoFixedUI, decimalFactorRat)
	owltoFixedWei := new(big.Int).Div(owltoFixedWeiRat.Num(), owltoFixedWeiRat.Denom())

	// 2. 计算 Owlto 百分比费用：amount * rate / 10000
	owltoRateWei := big.NewInt(0)
	if seg.OwltoFeeRateBps > 0 {
		owltoRateWei = new(big.Int).Mul(amtWei, big.NewInt(int64(seg.OwltoFeeRateBps)))
		owltoRateWei = new(big.Int).Div(owltoRateWei, big.NewInt(10000))
	}

	// 3. 总 Owlto 费用 = 固定 + 百分比
	result.OwltoFee = new(big.Int).Add(owltoFixedWei, owltoRateWei)

	// 4. Channel Fee（协议费用）这里根据实际聚合器 API 获取
	//    简化处理：暂时用 Owlto 固定费作为 Channel Fee（实际应调 API）
	result.ChannelFee = owltoFixedWei

	// 5. 总 Gas Fee
	result.TotalGasFee = new(big.Int).Add(result.ChannelFee, result.OwltoFee)

	// 6. Protocol Fee（如 Across 0.1%）
	if seg.ProtocolFeeRateBps > 0 {
		result.ProtocolFee = new(big.Int).Mul(amtWei, big.NewInt(int64(seg.ProtocolFeeRateBps)))
		result.ProtocolFee = new(big.Int).Div(result.ProtocolFee, big.NewInt(10000))
	} else {
		result.ProtocolFee = big.NewInt(0)
	}

	// 7. 转换为 UI 金额字符串（可选）
	result.TotalGasFeeUI = mgr.weiToUI(result.TotalGasFee, decimals)
	result.OwltoFeeUI = mgr.weiToUI(result.OwltoFee, decimals)
	result.ChannelFeeUI = mgr.weiToUI(result.ChannelFee, decimals)

	return result
}

// weiToUI 将最小单位转为 UI 金额字符串
func (mgr *AggregatorManager) weiToUI(wei *big.Int, decimals int32) string {
	if wei == nil || wei.Sign() == 0 {
		return "0"
	}
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	rat := new(big.Rat).SetFrac(wei, decimalFactor)
	return rat.FloatString(int(decimals))
}
