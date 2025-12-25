package request

type CreateSupplierRequest struct {
	SupplierName  string  `json:"supplier_name"`
	SupplierTitle *string `json:"supplier_title"`
	ContactNumber *string `json:"contact_number"`
	Email         *string `json:"email"`
	Address       *string `json:"address"`
	OtherContacts *string `json:"other_contacts"`
	SupplierType  string  `json:"supplier_type"`
	Country       *string `json:"country"`
	IsActive      *bool   `json:"is_active"`
}

type UpdateSupplierRequest struct {
	SupplierName  *string `json:"supplier_name"`
	SupplierTitle *string `json:"supplier_title"`
	ContactNumber *string `json:"contact_number"`
	Email         *string `json:"email"`
	Address       *string `json:"address"`
	OtherContacts *string `json:"other_contacts"`
	SupplierType  *string `json:"supplier_type"`
	Country       *string `json:"country"`
	IsActive      *bool   `json:"is_active"`
}
