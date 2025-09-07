package services

import (
	"car_service/dto/request"
	"car_service/entity"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"car_service/database"
)

type OrderService struct {
	db *database.Database
}

func NewOrderService(db *database.Database) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(req request.CreateOrderRequest) (*entity.CustomerOrder, error) {
	tx, err := s.db.Db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// First, create or get customer
	var customerID int64
	customerQuery := `
		INSERT INTO customers (customer_title, customer_name, contact_number, email, address, customer_type)
		VALUES (?, ?, ?, ?, ?, 'INDIVIDUAL')
		ON DUPLICATE KEY UPDATE
		customer_title = VALUES(customer_title),
		email = VALUES(email),
		address = VALUES(address),
		updated_at = CURRENT_TIMESTAMP
	`
	result, err := tx.Exec(customerQuery, req.CustomerTitle, req.CustomerName, req.ContactNumber, req.Email, req.Address)
	if err != nil {
		return nil, err
	}
	customerID, _ = result.LastInsertId()

	// If no insert occurred (duplicate), get the existing customer ID
	if customerID == 0 {
		err = tx.QueryRow("SELECT id FROM customers WHERE contact_number = ?", req.ContactNumber).Scan(&customerID)
		if err != nil {
			return nil, err
		}
	}

	// Generate order number
	orderNumber := fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), customerID)

	// Convert required features to JSON
	featuresJSON, _ := json.Marshal(req.RequiredFeatures)

	// Parse expected delivery date
	var expectedDelivery *time.Time
	if req.ExpectedDelivery != nil {
		if parsed, err := time.Parse("2006-01-02", *req.ExpectedDelivery); err == nil {
			expectedDelivery = &parsed
		}
	}

	// Create order
	orderQuery := `
		INSERT INTO customer_orders (
			order_number, customer_id, preferred_make, preferred_model, preferred_year_min, preferred_year_max,
			preferred_color, preferred_trim_level, max_mileage_km, min_auction_grade, required_features,
			order_type, expected_delivery_date, priority_level, preferred_port, shipping_method,
			include_insurance, budget_min, budget_max, payment_method, down_payment,
			special_requests, internal_notes, order_status, is_draft, order_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`

	orderStatus := "SUBMITTED"
	if req.IsDraft {
		orderStatus = "DRAFT"
	}

	result, err = tx.Exec(orderQuery, orderNumber, customerID, req.PreferredMake, req.PreferredModel,
		req.PreferredYear, req.PreferredYear, req.PreferredColor, req.TrimLevel, req.MaxMileage,
		req.MinAuctionGrade, string(featuresJSON), req.OrderType, expectedDelivery, req.Priority,
		req.PreferredPort, req.ShippingMethod, req.IncludeInsurance, req.BudgetMin, req.BudgetMax,
		req.PaymentMethod, req.DownPayment, req.SpecialRequests, req.InternalNotes, orderStatus, req.IsDraft)
	if err != nil {
		return nil, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Fetch and return the created order
	return s.GetOrderByID(orderID)
}

func (s *OrderService) GetOrderByID(id int64) (*entity.CustomerOrder, error) {
	query := `
		SELECT id, order_number, customer_id, preferred_make, preferred_model, preferred_year_min,
			preferred_year_max, preferred_color, preferred_trim_level, max_mileage_km, min_auction_grade,
			required_features, order_type, expected_delivery_date, priority_level, preferred_port,
			shipping_method, include_insurance, budget_min, budget_max, payment_method, down_payment,
			special_requests, internal_notes, order_status, is_draft, order_date, completed_date,
			created_at, updated_at
		FROM customer_orders WHERE id = ?
	`

	var order entity.CustomerOrder
	var featuresJSON string
	err := s.db.Db.QueryRow(query, id).Scan(
		&order.ID, &order.OrderNumber, &order.CustomerID, &order.PreferredMake, &order.PreferredModel,
		&order.PreferredYearMin, &order.PreferredYearMax, &order.PreferredColor, &order.PreferredTrimLevel,
		&order.MaxMileageKm, &order.MinAuctionGrade, &featuresJSON, &order.OrderType,
		&order.ExpectedDeliveryDate, &order.PriorityLevel, &order.PreferredPort, &order.ShippingMethod,
		&order.IncludeInsurance, &order.BudgetMin, &order.BudgetMax, &order.PaymentMethod,
		&order.DownPayment, &order.SpecialRequests, &order.InternalNotes, &order.OrderStatus,
		&order.IsDraft, &order.OrderDate, &order.CompletedDate, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse features JSON
	if featuresJSON != "" {
		json.Unmarshal([]byte(featuresJSON), &order.RequiredFeatures)
	}

	return &order, nil
}

func (s *OrderService) GetAllOrders(limit, offset int) ([]entity.CustomerOrder, error) {
	query := `
		SELECT co.id, co.order_number, co.customer_id, co.preferred_make, co.preferred_model,
			co.order_status, co.priority_level, co.budget_min, co.budget_max, co.order_date,
			co.created_at, c.customer_name, c.contact_number
		FROM customer_orders co
		LEFT JOIN customers c ON co.customer_id = c.id
		ORDER BY co.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entity.CustomerOrder
	for rows.Next() {
		var order entity.CustomerOrder
		var customerName, contactNumber sql.NullString
		err := rows.Scan(
			&order.ID, &order.OrderNumber, &order.CustomerID, &order.PreferredMake, &order.PreferredModel,
			&order.OrderStatus, &order.PriorityLevel, &order.BudgetMin, &order.BudgetMax, &order.OrderDate,
			&order.CreatedAt, &customerName, &contactNumber,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(orderID int64, status string) error {
	query := `UPDATE customer_orders SET order_status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Db.Exec(query, status, orderID)
	return err
}
