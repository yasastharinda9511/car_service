package entity

import "time"

type Customer struct {
	ID            int64     `json:"id" database:"id"`
	CustomerTitle *string   `json:"customer_title" database:"customer_title"`
	CustomerName  string    `json:"customer_name" database:"customer_name"`
	ContactNumber *string   `json:"contact_number" database:"contact_number"`
	Email         *string   `json:"email" database:"email"`
	Address       *string   `json:"address" database:"address"`
	OtherContacts *string   `json:"other_contacts" database:"other_contacts"`
	CustomerType  string    `json:"customer_type" database:"customer_type"`
	IsActive      bool      `json:"is_active" database:"is_active"`
	CreatedAt     time.Time `json:"created_at" database:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" database:"updated_at"`
}
