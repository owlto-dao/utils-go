package domain

type ReferralType string

const (
	Regular   ReferralType = "REGULAR"
	Kol       ReferralType = "KOL"
	Affiliate ReferralType = "AFFILIATE"
)

func IsValidReferralType(referralType string) bool {
	return string(Regular) == referralType || string(Kol) == referralType || string(Affiliate) == referralType
}

type CommissionRuleItem struct {
	MinTransactions int32  `json:"min_transactions"`
	MaxTransactions *int32 `json:"max_transactions"`
	CommissionRatio string `json:"commission_ratio"`
}

type CommissionRule struct {
	Items []*CommissionRuleItem `json:"items"`
}

type ReferralRebateRule struct {
	Items []*CommissionRuleItem `json:"items"`
}
