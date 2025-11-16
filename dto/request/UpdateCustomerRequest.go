package request

type UpdateCustomerRequest struct {
	CustomerTitle *string `json:"customer_title"` // Optional
	CustomerName  *string `json:"customer_name"`  // Optional
	ContactNumber *string `json:"contact_number"` // Optional
	Email         *string `json:"email"`          // Optional
	Address       *string `json:"address"`        // Optional
	OtherContacts *string `json:"other_contacts"` // Optional
	CustomerType  *string `json:"customer_type"`  // Optional
	IsActive      *bool   `json:"is_active"`      // Optional
}
