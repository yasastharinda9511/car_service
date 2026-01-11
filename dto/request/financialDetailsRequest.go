package request

import "encoding/json"

type FinancialDetailsRequest struct {
	ChargesLKR       *float64        `json:"charges_lkr"`
	TTLKR            *float64        `json:"tt_lkr"`
	DutyLKR          *float64        `json:"duty_lkr"`
	ClearingLKR      *float64        `json:"clearing_lkr"`
	OtherExpensesLKR json.RawMessage `json:"other_expenses_lkr"` // JSONB: {"car_wash": 1000, "detailing": 2500}
	TotalCostLKR     float64         `json:"total_cost_lkr"`     // Required field
}
