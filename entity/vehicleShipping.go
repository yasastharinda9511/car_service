package entity

import "time"

type VehicleShipping struct {
	ID               int64      `json:"id" database:"id"`
	VehicleID        int64      `json:"vehicle_id" database:"vehicle_id"`
	VesselName       *string    `json:"vessel_name" database:"vessel_name"`
	DepartureHarbour *string    `json:"departure_harbour" database:"departure_harbour"`
	ShipmentDate     *time.Time `json:"shipment_date" database:"shipment_date"`
	ArrivalDate      *time.Time `json:"arrival_date" database:"arrival_date"`
	ClearingDate     *time.Time `json:"clearing_date" database:"clearing_date"`
	ShippingStatus   string     `json:"shipping_status" database:"shipping_status"`
	CreatedAt        time.Time  `json:"created_at" database:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" database:"updated_at"`
}
