package entity

import "time"

type VehiclePurchase struct {
	ID              int64      `json:"id" database:"id"`
	VehicleID       int64      `json:"vehicle_id" database:"vehicle_id"`
	SupplierID      *int64     `json:"supplier_id" database:"supplier_id"`
	PurchaseRemarks *string    `json:"purchase_remarks" database:"purchase_remarks"`
	LCBank          *string    `json:"lc_bank" database:"lc_bank"`
	LCNumber        *string    `json:"lc_number" database:"lc_number"`
	LCCostJPY       *float64   `json:"lc_cost_jpy" database:"lc_cost_jpy"`
	PurchaseDate    *time.Time `json:"purchase_date" database:"purchase_date"`
	PurchaseStatus  string     `json:"purchase_status" database:"purchase_status"`
	CreatedAt       time.Time  `json:"created_at" database:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" database:"updated_at"`
}
