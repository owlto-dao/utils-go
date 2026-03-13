package loader

import (
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/owlto-dao/utils-go/alert"
)

// AggregatorID identifies an aggregator.
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

// AggregatorConfig stores base aggregator settings from t_aggregate_config.
type AggregatorConfig struct {
	ID         AggregatorID
	Name       string
	IsEnabled  bool
	Priority   int
	APIBaseURL string
}

// AggregateChainConfig stores per-chain aggregator settings from t_aggregate_chain_config.
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

// RouteConfig stores route settings from t_aggregator_route_config.
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
	MinAmount        string // UI amount
	MaxAmount        string // UI amount, "0" means no upper limit
	IsEnabled        bool
	Priority         int
}

// FeeSegment stores tiered fee settings from t_aggregator_route_fee_segment.
type FeeSegment struct {
	ID                 int64
	RouteID            int64  // References RouteConfig.ID
	MinAmountUI        string // Lower bound in UI amount
	MaxAmountUI        string // Upper bound in UI amount, "0" means no upper limit
	OwltoFeeFixedUI    string // Fixed Owlto surcharge in UI amount
	OwltoFeeRateBps    int    // Owlto percentage surcharge in basis points, 10 = 0.1%
	ProtocolFeeRateBps int    // Protocol fee rate in basis points
	IsEnabled          bool
	Priority           int
}

// FeeResult stores the computed fee result.
type FeeResult struct {
	BestAggregator     AggregatorID // Best aggregator
	ChannelFee         *big.Int     // Protocol fee in smallest unit
	OwltoFee           *big.Int     // Owlto surcharge in smallest unit
	TotalGasFee        *big.Int     // Total gas fee = ChannelFee + OwltoFee
	ProtocolFee        *big.Int     // Protocol fee amount, for example Across 0.1%
	ChannelFeeUI       string       // Protocol fee in UI amount
	OwltoFeeUI         string       // Owlto surcharge in UI amount
	TotalGasFeeUI      string       // Total fee in UI amount
	ProtocolFeeRateBps int          // Protocol fee rate in basis points
}

// RouteQueryResult stores the route query result.
type RouteQueryResult struct {
	Route       *RouteConfig
	FeeSegments []*FeeSegment
	IsAvailable bool
}

// SupportedChain stores chain information supported by aggregators.
type SupportedChain struct {
	ChainID   int64
	ChainName string
}

// BestRouteResult stores the best route selection result.
type BestRouteResult struct {
	Aggregator   AggregatorID
	Route        *RouteConfig
	FeeSegment   *FeeSegment
	FeeResult    *FeeResult
	IsAggregator bool // Whether the route uses an aggregator instead of Owlto Native
}

// AggregatorManager manages all aggregator-related configs in memory.
type AggregatorManager struct {
	// Aggregator configs
	aggregatorConfigs  map[AggregatorID]*AggregatorConfig
	chainConfigs       map[AggregatorID]map[int64]*AggregateChainConfig  // aggregateID -> chainID -> config
	chainConfigsByName map[AggregatorID]map[string]*AggregateChainConfig // aggregateID -> chainName(lower) -> config

	// Multi-dimensional route indexes
	routesByChainPair map[int64]map[int64][]*RouteConfig // fromChainID -> toChainID -> routes
	routesBySymbol    map[string][]*RouteConfig          // tokenSymbol(lower) -> routes
	routesByID        map[int64]*RouteConfig             // routeID -> route

	// Tiered fee indexes
	feeSegmentsByRouteID map[int64][]*FeeSegment // routeID -> segments, sorted by minAmount

	// Infrastructure
	db      *sql.DB
	alerter alert.Alerter
	mutex   *sync.RWMutex

	// Periodic refresh
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
		feeSegmentsByRouteID: make(map[int64][]*FeeSegment),
		db:                   db,
		alerter:              alerter,
		mutex:                new(sync.RWMutex),
		refreshInterval:      5 * time.Minute,
	}
}

// LoadAll loads all configs into memory.
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

// loadAggregatorConfigs loads aggregator configs from t_aggregate_config.
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

// loadAggregateChainConfigs loads aggregator chain configs from t_aggregate_chain_config.
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

// loadRouteConfigs loads routes from t_aggregator_route_config.
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

		// Index by chain pair.
		if _, ok := newByChainPair[r.FromChainID]; !ok {
			newByChainPair[r.FromChainID] = make(map[int64][]*RouteConfig)
		}
		newByChainPair[r.FromChainID][r.ToChainID] = append(newByChainPair[r.FromChainID][r.ToChainID], &r)

		// Index by token symbol.
		sym := strings.ToLower(r.FromTokenSymbol)
		newBySymbol[sym] = append(newBySymbol[sym], &r)

		// Index by route ID.
		newByID[r.ID] = &r
	}

	mgr.mutex.Lock()
	mgr.routesByChainPair = newByChainPair
	mgr.routesBySymbol = newBySymbol
	mgr.routesByID = newByID
	mgr.mutex.Unlock()
	return nil
}

// loadFeeSegments loads fee segments from t_aggregator_route_fee_segment.
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

	newByRouteID := make(map[int64][]*FeeSegment)
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

// GetAggregateChainConfigByChainID returns the chain config by aggregateID and chainID.
func (mgr *AggregatorManager) GetAggregateChainConfigByChainID(aggregateID AggregatorID, chainID int64) *AggregateChainConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if chainMap, ok := mgr.chainConfigs[aggregateID]; ok {
		return chainMap[chainID]
	}
	return nil
}

// GetAggregateChainConfigByChainName returns the chain config by aggregateID and chain name.
func (mgr *AggregatorManager) GetAggregateChainConfigByChainName(aggregateID AggregatorID, chainName string) *AggregateChainConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if chainMap, ok := mgr.chainConfigsByName[aggregateID]; ok {
		return chainMap[strings.ToLower(strings.TrimSpace(chainName))]
	}
	return nil
}

// GetAllRoutes returns all available routes.
func (mgr *AggregatorManager) GetAllRoutes() []*RouteConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	routes := make([]*RouteConfig, 0, len(mgr.routesByID))
	for _, route := range mgr.routesByID {
		routes = append(routes, route)
	}
	return routes
}

// GetAllSupportedChains returns all supported chains, deduplicated by chainID.
func (mgr *AggregatorManager) GetAllSupportedChains() []SupportedChain {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	chainByID := make(map[int64]SupportedChain)
	for _, chainMap := range mgr.chainConfigs {
		for chainID, cfg := range chainMap {
			if _, exists := chainByID[chainID]; exists {
				continue
			}
			chainByID[chainID] = SupportedChain{
				ChainID:   chainID,
				ChainName: cfg.ChainName,
			}
		}
	}

	chains := make([]SupportedChain, 0, len(chainByID))
	for _, chain := range chainByID {
		chains = append(chains, chain)
	}

	sort.Slice(chains, func(i, j int) bool {
		return chains[i].ChainID < chains[j].ChainID
	})
	return chains
}

// GetRoutesByChainPair returns all available routes for the given chain pair.
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

// GetRoutesByChainPairAndSymbol returns routes by chain pair and token symbol.
// ETH and WETH are treated as equivalent for all aggregators.
func (mgr *AggregatorManager) GetRoutesByChainPairAndSymbol(
	fromChainID, toChainID int64,
	tokenSymbol string,
) []*RouteConfig {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	var routes []*RouteConfig
	tokenUpper := strings.ToUpper(strings.TrimSpace(tokenSymbol))

	// Token mapping: requested token -> list of config tokens to match
	// ETH can match both ETH and WETH routes, but WETH only matches WETH
	tokenMappings := map[string][]string{
		"ETH": {"ETH", "WETH"},
	}

	// Get the list of tokens to match
	matchTokens := []string{tokenUpper}
	if mappings, ok := tokenMappings[tokenUpper]; ok {
		matchTokens = mappings
	}

	// Iterate routes and match any token
	if chainMap, ok := mgr.routesByChainPair[fromChainID]; ok {
		if rs, ok := chainMap[toChainID]; ok {
			for _, r := range rs {
				configTokenUpper := strings.ToUpper(r.FromTokenSymbol)
				for _, matchToken := range matchTokens {
					if configTokenUpper == matchToken {
						routes = append(routes, r)
						break
					}
				}
			}
		}
	}

	return routes
}

// GetFeeSegments returns the fee segments for a route.
func (mgr *AggregatorManager) GetFeeSegments(routeID int64) []*FeeSegment {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	return mgr.feeSegmentsByRouteID[routeID]
}

// FindBestRoute selects the best route based on amount.
// amountUI: UI amount such as "12.5" USDT
// tokenSymbol: token symbol such as "USDT" or "ETH"
// fromChainID, toChainID: chain IDs
// decimals: token decimals used to convert UI amount to the smallest unit
func (mgr *AggregatorManager) FindBestRoute(
	fromChainID, toChainID int64,
	tokenSymbol string,
	amountUI string,
	decimals int32,
) (*BestRouteResult, error) {
	// 1. Query all routes for the chain pair and token.
	routes := mgr.GetRoutesByChainPairAndSymbol(fromChainID, toChainID, tokenSymbol)
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found for %d->%d %s", fromChainID, toChainID, tokenSymbol)
	}

	// 2. Parse the UI amount.
	amtRat, ok := new(big.Rat).SetString(amountUI)
	if !ok || amtRat.Sign() <= 0 {
		return nil, fmt.Errorf("invalid amount: %s", amountUI)
	}

	// 3. Iterate through routes and find a matching fee segment.
	var bestResult *BestRouteResult
	var lowestFee *big.Rat

	for _, route := range routes {
		// Check whether the amount falls within the route min/max range.
		if !mgr.isAmountInRouteRange(amtRat, route) {
			continue
		}

		// Find the matching fee segment for this route.
		segment := mgr.findMatchingSegment(route.ID, amtRat)
		if segment == nil {
			continue // No matching fee segment.
		}

		// Calculate fees for the matched segment.
		feeResult := mgr.calculateFee(amtRat, segment, decimals)
		if feeResult == nil {
			continue
		}
		feeResult.BestAggregator = route.AggregateID

		// Keep the route with the lowest fee.
		feeRat := new(big.Rat).SetInt(feeResult.TotalGasFee)
		if lowestFee == nil || feeRat.Cmp(lowestFee) < 0 {
			lowestFee = feeRat
			bestResult = &BestRouteResult{
				Aggregator:   route.AggregateID,
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

// isAmountInRouteRange checks whether the amount is within the route min/max range.
func (mgr *AggregatorManager) isAmountInRouteRange(amtRat *big.Rat, route *RouteConfig) bool {
	// Parse min.
	minRat := new(big.Rat)
	if route.MinAmount != "" && route.MinAmount != "0" {
		if _, ok := minRat.SetString(route.MinAmount); !ok {
			return false
		}
		if amtRat.Cmp(minRat) < 0 {
			return false
		}
	}

	// Parse max. "0" or empty means no upper limit.
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

// findMatchingSegment finds the fee segment matching the given amount.
func (mgr *AggregatorManager) findMatchingSegment(routeID int64, amtRat *big.Rat) *FeeSegment {
	segments := mgr.GetFeeSegments(routeID)
	if segments == nil {
		return nil
	}

	for _, seg := range segments {
		// Parse min.
		minRat := new(big.Rat)
		if seg.MinAmountUI != "" {
			if _, ok := minRat.SetString(seg.MinAmountUI); !ok {
				continue
			}
		}

		// Parse max.
		maxRat := new(big.Rat)
		unlimited := seg.MaxAmountUI == "" || seg.MaxAmountUI == "0"
		if !unlimited {
			if _, ok := maxRat.SetString(seg.MaxAmountUI); !ok {
				continue
			}
		}

		// Check whether amt is within the [min, max) range.
		if amtRat.Cmp(minRat) >= 0 {
			if unlimited || amtRat.Cmp(maxRat) < 0 {
				return seg
			}
		}
	}

	return nil
}

// calculateFee calculates fees from a fee segment.
// amtRat: UI amount as big.Rat
// seg: matched fee segment
// decimals: token decimals
func (mgr *AggregatorManager) calculateFee(amtRat *big.Rat, seg *FeeSegment, decimals int32) *FeeResult {
	result := &FeeResult{
		BestAggregator:     AggregatorNone, // Set by the caller.
		ProtocolFeeRateBps: seg.ProtocolFeeRateBps,
	}

	// Compute the precision factor: 10^decimals.
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	decimalFactorRat := new(big.Rat).SetInt(decimalFactor)

	// Convert the UI amount to the smallest unit.
	amtWeiRat := new(big.Rat).Mul(amtRat, decimalFactorRat)
	amtWei := new(big.Int).Div(amtWeiRat.Num(), amtWeiRat.Denom())

	// 1. Calculate the fixed Owlto fee.
	owltoFixedUI := new(big.Rat)
	if seg.OwltoFeeFixedUI != "" && seg.OwltoFeeFixedUI != "0" {
		owltoFixedUI.SetString(seg.OwltoFeeFixedUI)
	}
	owltoFixedWeiRat := new(big.Rat).Mul(owltoFixedUI, decimalFactorRat)
	owltoFixedWei := new(big.Int).Div(owltoFixedWeiRat.Num(), owltoFixedWeiRat.Denom())

	// 2. Calculate the Owlto percentage fee: amount * rate / 10000.
	owltoRateWei := big.NewInt(0)
	if seg.OwltoFeeRateBps > 0 {
		owltoRateWei = new(big.Int).Mul(amtWei, big.NewInt(int64(seg.OwltoFeeRateBps)))
		owltoRateWei = new(big.Int).Div(owltoRateWei, big.NewInt(10000))
	}

	// 3. Total Owlto fee = fixed + percentage.
	result.OwltoFee = new(big.Int).Add(owltoFixedWei, owltoRateWei)

	// 4. Channel fee should come from the actual aggregator API.
	//    For now, use the Owlto fixed fee as a placeholder channel fee.
	result.ChannelFee = owltoFixedWei

	// 5. Total gas fee.
	result.TotalGasFee = new(big.Int).Add(result.ChannelFee, result.OwltoFee)

	// 6. Protocol fee, for example Across 0.1%.
	if seg.ProtocolFeeRateBps > 0 {
		result.ProtocolFee = new(big.Int).Mul(amtWei, big.NewInt(int64(seg.ProtocolFeeRateBps)))
		result.ProtocolFee = new(big.Int).Div(result.ProtocolFee, big.NewInt(10000))
	} else {
		result.ProtocolFee = big.NewInt(0)
	}

	// 7. Convert results back to UI amount strings.
	result.TotalGasFeeUI = mgr.weiToUI(result.TotalGasFee, decimals)
	result.OwltoFeeUI = mgr.weiToUI(result.OwltoFee, decimals)
	result.ChannelFeeUI = mgr.weiToUI(result.ChannelFee, decimals)

	return result
}

// weiToUI converts the smallest unit to a UI amount string.
func (mgr *AggregatorManager) weiToUI(wei *big.Int, decimals int32) string {
	if wei == nil || wei.Sign() == 0 {
		return "0"
	}
	decimalFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	rat := new(big.Rat).SetFrac(wei, decimalFactor)
	return rat.FloatString(int(decimals))
}
