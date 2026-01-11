package request

import "time"

type PurchaseRequest struct {
	SupplierID      *int64     `json:"supplier_id"`
	PurchaseRemarks *string    `json:"purchase_remarks"`
	LCBank          *string    `json:"lc_bank"`
	LCNumber        *string    `json:"lc_number"`
	LCCostJPY       *float64   `json:"lc_cost_jpy"`
	ExchangeRate    *float64   `json:"exchange_rate"`
	PurchaseDate    *time.Time `json:"purchase_date"` // "2024-01-15T10:30:00Z"
	PurchaseStatus  *string    `json:"purchase_status"`
}
