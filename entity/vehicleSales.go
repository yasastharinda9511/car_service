package entity

import "time"

type VehicleSales struct {
	ID              int64      `json:"id" database:"id"`
	VehicleID       int64      `json:"vehicle_id" database:"vehicle_id"`
	SoldDate        *time.Time `json:"sold_date" database:"sold_date"`
	Revenue         *float64   `json:"revenue" database:"revenue"`
	Profit          *float64   `json:"profit" database:"profit"`
	SoldToName      *string    `json:"sold_to_name" database:"sold_to_name"`
	SoldToTitle     *string    `json:"sold_to_title" database:"sold_to_title"`
	ContactNumber   *string    `json:"contact_number" database:"contact_number"`
	CustomerAddress *string    `json:"customer_address" database:"customer_address"`
	OtherContacts   *string    `json:"other_contacts" database:"other_contacts"`
	SaleRemarks     *string    `json:"sale_remarks" database:"sale_remarks"`
	SaleStatus      string     `json:"sale_status" database:"sale_status"`
	CreatedAt       time.Time  `json:"created_at" database:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" database:"updated_at"`
}
