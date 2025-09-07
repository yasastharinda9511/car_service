package entity

import "time"

type VehiclePurchase struct {
	ID                      int64      `json:"id" database:"id"`
	VehicleID               int64      `json:"vehicle_id" database:"vehicle_id"`
	BoughtFromName          *string    `json:"bought_from_name" database:"bought_from_name"`
	BoughtFromTitle         *string    `json:"bought_from_title" database:"bought_from_title"`
	BoughtFromContact       *string    `json:"bought_from_contact" database:"bought_from_contact"`
	BoughtFromAddress       *string    `json:"bought_from_address" database:"bought_from_address"`
	BoughtFromOtherContacts *string    `json:"bought_from_other_contacts" database:"bought_from_other_contacts"`
	PurchaseRemarks         *string    `json:"purchase_remarks" database:"purchase_remarks"`
	LCBank                  *string    `json:"lc_bank" database:"lc_bank"`
	LCNumber                *string    `json:"lc_number" database:"lc_number"`
	LCCostJPY               *float64   `json:"lc_cost_jpy" database:"lc_cost_jpy"`
	PurchaseDate            *time.Time `json:"purchase_date" database:"purchase_date"`
	CreatedAt               time.Time  `json:"created_at" database:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" database:"updated_at"`
}
