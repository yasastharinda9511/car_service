package request

type FinancialDetailsRequest struct {
	ChargesLKR       *float64 `json:"charges_lkr"`
	TTLKR            *float64 `json:"tt_lkr"`
	DutyLKR          *float64 `json:"duty_lkr"`
	ClearingLKR      *float64 `json:"clearing_lkr"`
	OtherExpensesLKR *float64 `json:"other_expenses_lkr"`
	TotalCostLKR     float64  `json:"total_cost_lkr"` // Required field
}
