package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"context"
	"fmt"
	"strings"
)

type CustomerRepository struct{}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{}
}

// CreateCustomer creates a new customer
func (r *CustomerRepository) CreateCustomer(ctx context.Context, exec database.Executor, req request.CreateCustomerRequest) (*entity.Customer, error) {
	query := `
        INSERT INTO cars.customers (customer_title, customer_name, contact_number, email, address, other_contacts, customer_type, is_active)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, customer_title, customer_name, contact_number, email, address, other_contacts, customer_type, is_active, created_at, updated_at
    `

	var customer entity.Customer
	err := exec.QueryRowContext(ctx, query,
		req.CustomerTitle, req.CustomerName, req.ContactNumber, req.Email,
		req.Address, req.OtherContacts, req.CustomerType, req.IsActive,
	).Scan(
		&customer.ID, &customer.CustomerTitle, &customer.CustomerName,
		&customer.ContactNumber, &customer.Email, &customer.Address,
		&customer.OtherContacts, &customer.CustomerType, &customer.IsActive,
		&customer.CreatedAt, &customer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// GetAllCustomers retrieves all customers with optional filtering, search, and pagination
func (r *CustomerRepository) GetAllCustomers(ctx context.Context, exec database.Executor, limit, offset int, customerType *string, activeOnly bool, searchTerm string) ([]entity.Customer, error) {
	query := `
        SELECT id, customer_title, customer_name, contact_number, email, address,
               other_contacts, customer_type, is_active, created_at, updated_at
        FROM cars.customers
    `

	var conditions []string
	var args []interface{}
	argCount := 1

	// Add search condition if search term is provided
	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(customer_name) LIKE LOWER($%d) OR LOWER(contact_number) LIKE LOWER($%d) OR LOWER(email) LIKE LOWER($%d) OR LOWER(customer_type::text) LIKE LOWER($%d))", argCount, argCount, argCount, argCount))
		args = append(args, searchPattern)
		argCount++
	}

	if customerType != nil && *customerType != "" {
		conditions = append(conditions, fmt.Sprintf("customer_type = $%d", argCount))
		args = append(args, *customerType)
		argCount++
	}

	if activeOnly {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, true)
		argCount++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY customer_name"

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []entity.Customer
	for rows.Next() {
		var customer entity.Customer
		err := rows.Scan(
			&customer.ID, &customer.CustomerTitle, &customer.CustomerName,
			&customer.ContactNumber, &customer.Email, &customer.Address,
			&customer.OtherContacts, &customer.CustomerType, &customer.IsActive,
			&customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}

// GetAllCustomersCount retrieves the total count of customers with optional filtering and search
func (r *CustomerRepository) GetAllCustomersCount(ctx context.Context, exec database.Executor, customerType *string, activeOnly bool, searchTerm string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM cars.customers`

	var conditions []string
	var args []interface{}
	argCount := 1

	// Add search condition if search term is provided
	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(customer_name) LIKE LOWER($%d) OR LOWER(contact_number) LIKE LOWER($%d) OR LOWER(email) LIKE LOWER($%d) OR LOWER(customer_type::text) LIKE LOWER($%d))", argCount, argCount, argCount, argCount))
		args = append(args, searchPattern)
		argCount++
	}

	if customerType != nil && *customerType != "" {
		conditions = append(conditions, fmt.Sprintf("customer_type = $%d", argCount))
		args = append(args, *customerType)
		argCount++
	}

	if activeOnly {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, true)
		argCount++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	err := exec.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetCustomerByID retrieves a customer by ID
func (r *CustomerRepository) GetCustomerByID(ctx context.Context, exec database.Executor, id int64) (*entity.Customer, error) {
	query := `
        SELECT id, customer_title, customer_name, contact_number, email, address,
               other_contacts, customer_type, is_active, created_at, updated_at
        FROM cars.customers
        WHERE id = $1
    `

	var customer entity.Customer
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.CustomerTitle, &customer.CustomerName,
		&customer.ContactNumber, &customer.Email, &customer.Address,
		&customer.OtherContacts, &customer.CustomerType, &customer.IsActive,
		&customer.CreatedAt, &customer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// UpdateCustomer updates a customer's information
func (r *CustomerRepository) UpdateCustomer(ctx context.Context, exec database.Executor, id int64, req request.UpdateCustomerRequest) error {
	query := `
        UPDATE cars.customers
        SET customer_title = COALESCE($2, customer_title),
            customer_name = COALESCE($3, customer_name),
            contact_number = COALESCE($4, contact_number),
            email = COALESCE($5, email),
            address = COALESCE($6, address),
            other_contacts = COALESCE($7, other_contacts),
            customer_type = COALESCE($8, customer_type),
            is_active = COALESCE($9, is_active),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `

	_, err := exec.ExecContext(ctx, query, id,
		req.CustomerTitle, req.CustomerName, req.ContactNumber, req.Email,
		req.Address, req.OtherContacts, req.CustomerType, req.IsActive,
	)
	return err
}

// DeleteCustomer soft deletes a customer by setting is_active to false
func (r *CustomerRepository) DeleteCustomer(ctx context.Context, exec database.Executor, id int64) error {
	query := `
        UPDATE cars.customers
        SET is_active = false,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `

	_, err := exec.ExecContext(ctx, query, id)
	return err
}

// SearchCustomers searches customers by name, contact, or email
func (r *CustomerRepository) SearchCustomers(ctx context.Context, exec database.Executor, searchTerm string) ([]entity.Customer, error) {
	query := `
        SELECT id, customer_title, customer_name, contact_number, email, address,
               other_contacts, customer_type, is_active, created_at, updated_at
        FROM cars.customers
        WHERE (
            LOWER(customer_name) LIKE LOWER($1) OR
            LOWER(contact_number) LIKE LOWER($1) OR
            LOWER(email) LIKE LOWER($1)
        )
        AND is_active = true
        ORDER BY customer_name
        LIMIT 50
    `

	searchPattern := "%" + searchTerm + "%"
	rows, err := exec.QueryContext(ctx, query, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []entity.Customer
	for rows.Next() {
		var customer entity.Customer
		err := rows.Scan(
			&customer.ID, &customer.CustomerTitle, &customer.CustomerName,
			&customer.ContactNumber, &customer.Email, &customer.Address,
			&customer.OtherContacts, &customer.CustomerType, &customer.IsActive,
			&customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}
