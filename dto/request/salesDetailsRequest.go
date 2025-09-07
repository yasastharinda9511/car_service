package request

type SalesDetailsRequest struct {
	SoldDate        *string  `json:"sold_date"` // "2024-01-15T10:30:00Z"
	Revenue         *float64 `json:"revenue"`
	Profit          *float64 `json:"profit"`
	SoldToName      *string  `json:"sold_to_name"`
	SoldToTitle     *string  `json:"sold_to_title"`
	ContactNumber   *string  `json:"contact_number"`
	CustomerAddress *string  `json:"customer_address"`
	OtherContacts   *string  `json:"other_contacts"`
	SaleRemarks     *string  `json:"sale_remarks"`
	SaleStatus      string   `json:"sale_status"` // Required field
}
