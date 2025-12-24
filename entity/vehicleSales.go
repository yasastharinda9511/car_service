package entity

import "time"

type VehicleSales struct {
	ID           int64      `json:"id" database:"id"`
	VehicleID    int64      `json:"vehicle_id" database:"vehicle_id"`
	CustomerID   *int64     `json:"customer_id" database:"customer_id"`
	CustomerName string     `json:"customer_name,omitempty"`
	SoldDate     *time.Time `json:"sold_date" database:"sold_date"`
	Revenue      *float64   `json:"revenue" database:"revenue"`
	Profit       *float64   `json:"profit" database:"profit"`
	SaleRemarks  *string    `json:"sale_remarks" database:"sale_remarks"`
	SaleStatus   string     `json:"sale_status" database:"sale_status"`
	CreatedAt    time.Time  `json:"created_at" database:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" database:"updated_at"`
}
