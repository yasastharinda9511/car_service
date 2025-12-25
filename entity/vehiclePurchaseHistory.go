package entity

import "time"

type VehiclePurchaseHistory struct {
	ID              int64      `json:"id"`
	VehicleID       int64      `json:"vehicle_id"`
	OldStatus       *string    `json:"old_status"`
	NewStatus       string     `json:"new_status"`
	SupplierID      *int64     `json:"supplier_id"`
	LCBank          *string    `json:"lc_bank"`
	LCNumber        *string    `json:"lc_number"`
	LCCostJPY       *float64   `json:"lc_cost_jpy"`
	PurchaseDate    *time.Time `json:"purchase_date"`
	PurchaseRemarks *string    `json:"purchase_remarks"`
	ChangedBy       *string    `json:"changed_by"`
	ChangeRemarks   *string    `json:"change_remarks"`
	ChangedAt       time.Time  `json:"changed_at"`
}

type VehiclePurchaseHistoryWithDetails struct {
	VehiclePurchaseHistory
	VehicleCode           string   `json:"vehicle_code"`
	Make                  string   `json:"make"`
	Model                 string   `json:"model"`
	ChassisID             string   `json:"chassis_id"`
	HoursInPreviousStatus *float64 `json:"hours_in_previous_status"`
}
