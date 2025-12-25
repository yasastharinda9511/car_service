package repository

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/entity"
	"context"
	"fmt"
	"strings"
)

type SupplierRepository struct{}

func NewSupplierRepository() *SupplierRepository {
	return &SupplierRepository{}
}

// CreateSupplier creates a new supplier
func (r *SupplierRepository) CreateSupplier(ctx context.Context, exec database.Executor, req request.CreateSupplierRequest) (*entity.Supplier, error) {
	// Set default country if not provided
	country := "Japan"
	if req.Country != nil && *req.Country != "" {
		country = *req.Country
	}

	query := `
        INSERT INTO cars.suppliers (supplier_name, supplier_title, contact_number, email, address, other_contacts, supplier_type, country, is_active)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, supplier_name, supplier_title, contact_number, email, address, other_contacts, supplier_type, country, is_active, created_at, updated_at
    `

	var supplier entity.Supplier
	err := exec.QueryRowContext(ctx, query,
		req.SupplierName, req.SupplierTitle, req.ContactNumber, req.Email,
		req.Address, req.OtherContacts, req.SupplierType, country, req.IsActive,
	).Scan(
		&supplier.ID, &supplier.SupplierName, &supplier.SupplierTitle,
		&supplier.ContactNumber, &supplier.Email, &supplier.Address,
		&supplier.OtherContacts, &supplier.SupplierType, &supplier.Country,
		&supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

// GetAllSuppliers retrieves all suppliers with optional filtering
func (r *SupplierRepository) GetAllSuppliers(ctx context.Context, exec database.Executor, supplierType *string, activeOnly bool) ([]entity.Supplier, error) {
	query := `
        SELECT id, supplier_name, supplier_title, contact_number, email, address,
               other_contacts, supplier_type, country, is_active, created_at, updated_at
        FROM cars.suppliers
    `

	var conditions []string
	var args []interface{}
	argCount := 1

	if supplierType != nil && *supplierType != "" {
		conditions = append(conditions, fmt.Sprintf("supplier_type = $%d", argCount))
		args = append(args, *supplierType)
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

	query += " ORDER BY supplier_name"

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []entity.Supplier
	for rows.Next() {
		var supplier entity.Supplier
		err := rows.Scan(
			&supplier.ID, &supplier.SupplierName, &supplier.SupplierTitle,
			&supplier.ContactNumber, &supplier.Email, &supplier.Address,
			&supplier.OtherContacts, &supplier.SupplierType, &supplier.Country,
			&supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}

// GetSupplierByID retrieves a supplier by ID
func (r *SupplierRepository) GetSupplierByID(ctx context.Context, exec database.Executor, id int64) (*entity.Supplier, error) {
	query := `
        SELECT id, supplier_name, supplier_title, contact_number, email, address,
               other_contacts, supplier_type, country, is_active, created_at, updated_at
        FROM cars.suppliers
        WHERE id = $1
    `

	var supplier entity.Supplier
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&supplier.ID, &supplier.SupplierName, &supplier.SupplierTitle,
		&supplier.ContactNumber, &supplier.Email, &supplier.Address,
		&supplier.OtherContacts, &supplier.SupplierType, &supplier.Country,
		&supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

// UpdateSupplier updates a supplier's information
func (r *SupplierRepository) UpdateSupplier(ctx context.Context, exec database.Executor, id int64, req request.UpdateSupplierRequest) error {
	query := `
        UPDATE cars.suppliers
        SET supplier_name = COALESCE($2, supplier_name),
            supplier_title = COALESCE($3, supplier_title),
            contact_number = COALESCE($4, contact_number),
            email = COALESCE($5, email),
            address = COALESCE($6, address),
            other_contacts = COALESCE($7, other_contacts),
            supplier_type = COALESCE($8, supplier_type),
            country = COALESCE($9, country),
            is_active = COALESCE($10, is_active),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `

	_, err := exec.ExecContext(ctx, query, id,
		req.SupplierName, req.SupplierTitle, req.ContactNumber, req.Email,
		req.Address, req.OtherContacts, req.SupplierType, req.Country, req.IsActive,
	)
	return err
}

// DeleteSupplier soft deletes a supplier by setting is_active to false
func (r *SupplierRepository) DeleteSupplier(ctx context.Context, exec database.Executor, id int64) error {
	query := `
        UPDATE cars.suppliers
        SET is_active = false,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `

	_, err := exec.ExecContext(ctx, query, id)
	return err
}

// SearchSuppliers searches suppliers by name, contact, or email
func (r *SupplierRepository) SearchSuppliers(ctx context.Context, exec database.Executor, searchTerm string) ([]entity.Supplier, error) {
	query := `
        SELECT id, supplier_name, supplier_title, contact_number, email, address,
               other_contacts, supplier_type, country, is_active, created_at, updated_at
        FROM cars.suppliers
        WHERE (
            LOWER(supplier_name) LIKE LOWER($1) OR
            LOWER(contact_number) LIKE LOWER($1) OR
            LOWER(email) LIKE LOWER($1)
        )
        AND is_active = true
        ORDER BY supplier_name
        LIMIT 50
    `

	searchPattern := "%" + searchTerm + "%"
	rows, err := exec.QueryContext(ctx, query, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []entity.Supplier
	for rows.Next() {
		var supplier entity.Supplier
		err := rows.Scan(
			&supplier.ID, &supplier.SupplierName, &supplier.SupplierTitle,
			&supplier.ContactNumber, &supplier.Email, &supplier.Address,
			&supplier.OtherContacts, &supplier.SupplierType, &supplier.Country,
			&supplier.IsActive, &supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}
