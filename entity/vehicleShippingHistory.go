package entity

import "time"

type VehicleShippingHistory struct {
	ID               int64      `json:"id"`
	VehicleID        int64      `json:"vehicle_id"`
	OldStatus        *string    `json:"old_status"`
	NewStatus        string     `json:"new_status"`
	VesselName       *string    `json:"vessel_name"`
	DepartureHarbour *string    `json:"departure_harbour"`
	ShipmentDate     *time.Time `json:"shipment_date"`
	ArrivalDate      *time.Time `json:"arrival_date"`
	ClearingDate     *time.Time `json:"clearing_date"`
	ChangedBy        *string    `json:"changed_by"`
	ChangeRemarks    *string    `json:"change_remarks"`
	ChangedAt        time.Time  `json:"changed_at"`
}

// VehicleShippingHistoryWithDetails includes vehicle details for better context
type VehicleShippingHistoryWithDetails struct {
	VehicleShippingHistory
	VehicleCode           string   `json:"vehicle_code"`
	Make                  string   `json:"make"`
	Model                 string   `json:"model"`
	ChassisID             string   `json:"chassis_id"`
	HoursInPreviousStatus *float64 `json:"hours_in_previous_status"`
}
