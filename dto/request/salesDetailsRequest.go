package request

type SalesDetailsRequest struct {
	CustomerID      *int64   `json:"customer_id"` // Optional: Link to customer record
	SoldDate        *string  `json:"sold_date"`   // "2024-01-15T10:30:00Z"
	Revenue         *float64 `json:"revenue"`
	Profit          *float64 `json:"profit"`           // Auto-calculated as revenue - total_cost, ignored if provided
	SoldToName      *string  `json:"sold_to_name"`     // Legacy/backup field
	SoldToTitle     *string  `json:"sold_to_title"`    // Legacy/backup field
	ContactNumber   *string  `json:"contact_number"`   // Legacy/backup field
	CustomerAddress *string  `json:"customer_address"` // Legacy/backup field
	OtherContacts   *string  `json:"other_contacts"`   // Legacy/backup field
	SaleRemarks     *string  `json:"sale_remarks"`
	SaleStatus      string   `json:"sale_status"` // Required field
}
