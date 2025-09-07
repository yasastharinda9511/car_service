package request

type PurchaseRequest struct {
	BoughtFromName          *string  `json:"bought_from_name"`
	BoughtFromTitle         *string  `json:"bought_from_title"`
	BoughtFromContact       *string  `json:"bought_from_contact"`
	BoughtFromAddress       *string  `json:"bought_from_address"`
	BoughtFromOtherContacts *string  `json:"bought_from_other_contacts"`
	PurchaseRemarks         *string  `json:"purchase_remarks"`
	LCBank                  *string  `json:"lc_bank"`
	LCNumber                *string  `json:"lc_number"`
	LCCostJPY               *float64 `json:"lc_cost_jpy"`
	PurchaseDate            *string  `json:"purchase_date"` // "2024-01-15T10:30:00Z"
}
