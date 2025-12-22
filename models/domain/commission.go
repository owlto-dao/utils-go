package domain

type ReferralType string

const (
	Normal ReferralType = "NORMAL"
	Kol    ReferralType = "KOL"
)

func IsValidReferralType(referralType string) bool {
	return string(Normal) == referralType || string(Kol) == referralType
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
