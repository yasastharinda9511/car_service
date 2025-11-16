package request

type CreateCustomerRequest struct {
	CustomerTitle *string `json:"customer_title"` // Optional: Mr, Ms, Dr, etc.
	CustomerName  string  `json:"customer_name"`  // Required
	ContactNumber *string `json:"contact_number"` // Optional
	Email         *string `json:"email"`          // Optional
	Address       *string `json:"address"`        // Optional
	OtherContacts *string `json:"other_contacts"` // Optional
	CustomerType  string  `json:"customer_type"`  // Required: INDIVIDUAL or BUSINESS
	IsActive      *bool   `json:"is_active"`      // Optional, defaults to true
}
