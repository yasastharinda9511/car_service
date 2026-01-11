package entity

import (
	"encoding/json"
	"time"
)

type VehicleFinancials struct {
	ID               int64           `json:"id" database:"id"`
	VehicleID        int64           `json:"vehicle_id" database:"vehicle_id"`
	ChargesLKR       *float64        `json:"charges_lkr" database:"charges_lkr"`
	TTLKR            *float64        `json:"tt_lkr" database:"tt_lkr"`
	DutyLKR          *float64        `json:"duty_lkr" database:"duty_lkr"`
	ClearingLKR      *float64        `json:"clearing_lkr" database:"clearing_lkr"`
	OtherExpensesLKR json.RawMessage `json:"other_expenses_lkr" database:"other_expenses_lkr"`
	TotalCostLKR     float64         `json:"total_cost_lkr" database:"total_cost_lkr"`
	CreatedAt        time.Time       `json:"created_at" database:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" database:"updated_at"`
}
