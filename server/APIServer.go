package server

import (
	"car_service/database"
	"car_service/dto/request"
	"car_service/services"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type APIServer struct {
	vehicleService *services.VehicleService
	orderService   *services.OrderService
	router         *mux.Router
}

func NewAPIServer(db *database.Database) *APIServer {
	server := &APIServer{
		vehicleService: services.NewVehicleService(db),
		orderService:   services.NewOrderService(db),
		router:         mux.NewRouter(),
	}

	server.setupRoutes()
	return server
}

// CORS Middleware
func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) setupRoutes() {
	// Apply CORS middleware
	s.router.Use(s.corsMiddleware)

	// Health check
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")

	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Vehicle routes
	vehicles := api.PathPrefix("/vehicles").Subrouter()
	vehicles.HandleFunc("", s.getVehicles).Methods("GET")
	vehicles.HandleFunc("/{id}", s.getVehicle).Methods("GET")
	vehicles.HandleFunc("", s.createVehicle).Methods("POST")

	vehicles.HandleFunc("/{id}/shipping", s.updateShipping).Methods("PUT")
	vehicles.HandleFunc("/{id}/purchase", s.updatePurchase).Methods("PUT")
	vehicles.HandleFunc("/{id}/financials", s.updateFinancials).Methods("PUT")
	vehicles.HandleFunc("/{id}/sales", s.updateSales).Methods("PUT")
	vehicles.HandleFunc("/{id}", s.updateVehicle).Methods("PUT")

	// Order routes
	orders := api.PathPrefix("/orders").Subrouter()
	orders.HandleFunc("", s.getOrders).Methods("GET")
	orders.HandleFunc("/{id}", s.getOrder).Methods("GET")
	orders.HandleFunc("", s.createOrder).Methods("POST")
	orders.HandleFunc("/{id}/status", s.updateOrderStatus).Methods("PUT")

	// Analytics routes
	analytics := api.PathPrefix("/analytics").Subrouter()
	analytics.HandleFunc("/dashboard", s.getDashboardStats).Methods("GET")
	analytics.HandleFunc("/sales-summary", s.getSalesSummary).Methods("GET")
	analytics.HandleFunc("/inventory-status", s.getInventoryStatus).Methods("GET")

	// Vehicle Makes routes
	makes := api.PathPrefix("/makes").Subrouter()
	makes.HandleFunc("", s.createVehicleMake).Methods("POST")
	makes.HandleFunc("", s.getVehicleMakes).Methods("GET")
	makes.HandleFunc("/{id}", s.updateVehicleMake).Methods("PUT")

	// Vehicle Models routes
	models := api.PathPrefix("/models").Subrouter()
	models.HandleFunc("", s.createVehicleModel).Methods("POST")
	models.HandleFunc("", s.getVehicleModels).Methods("GET")
	models.HandleFunc("/{id}", s.getVehicleModel).Methods("GET")
	models.HandleFunc("/{id}", s.updateVehicleModel).Methods("PUT")
}

// Helper functions for JSON responses
func (s *APIServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *APIServer) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// Health check handler
func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "car-deals-api",
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Vehicle handlers
func (s *APIServer) getVehicles(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	vehicles, err := s.vehicleService.GetAllVehicles(limit, offset)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data": vehicles,
		"meta": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": len(vehicles),
		},
	}
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) getVehicle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	vehicle, err := s.vehicleService.GetVehicleByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": vehicle})
}

func (s *APIServer) createVehicle(w http.ResponseWriter, r *http.Request) {
	var req request.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.Make == "" || req.Model == "" || req.ChassisID == "" {
		s.writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	vehicle, err := s.vehicleService.CreateVehicle(req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{"data": vehicle})
}

func (s *APIServer) updateShipping(w http.ResponseWriter, r *http.Request) {
	var req request.ShippingDetailsRequest

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = s.vehicleService.UpdateShippingStatus(id, req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Shipping status updated successfully"})
}

func (s *APIServer) updatePurchase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req request.PurchaseRequest
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Parse purchase date if provided
	var purchaseDate *time.Time
	if req.PurchaseDate != nil && *req.PurchaseDate != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.PurchaseDate); err == nil {
			purchaseDate = &parsed
		} else {
			s.writeError(w, http.StatusBadRequest, "Invalid purchase_date format. Use RFC3339 format")
			return
		}
	}

	err = s.vehicleService.UpdatePurchaseDetails(id, req.BoughtFromName, req.BoughtFromTitle,
		req.BoughtFromContact, req.BoughtFromAddress, req.BoughtFromOtherContacts,
		req.PurchaseRemarks, req.LCBank, req.LCNumber, req.LCCostJPY, purchaseDate)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Purchase details updated successfully"})
}

func (s *APIServer) updateFinancials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var req request.FinancialDetailsRequest
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required field
	if req.TotalCostLKR <= 0 {
		s.writeError(w, http.StatusBadRequest, "Total cost must be greater than 0")
		return
	}

	err = s.vehicleService.UpdateFinancialDetails(id, req.ChargesLKR, req.TTLKR,
		req.DutyLKR, req.ClearingLKR, req.OtherExpensesLKR, req.TotalCostLKR)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Financial details updated successfully"})
}

func (s *APIServer) updateSales(w http.ResponseWriter, r *http.Request) {
	var req request.SalesDetailsRequest
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required field
	if req.SaleStatus == "" {
		s.writeError(w, http.StatusBadRequest, "Sale status is required")
		return
	}

	// Validate sale status enum values
	validStatuses := map[string]bool{
		"AVAILABLE": true, "RESERVED": true, "SOLD": true, "CANCELLED": true,
	}
	if !validStatuses[req.SaleStatus] {
		s.writeError(w, http.StatusBadRequest, "Invalid sale status")
		return
	}

	// Parse sold date if provided
	var soldDate *time.Time
	if req.SoldDate != nil && *req.SoldDate != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.SoldDate); err == nil {
			soldDate = &parsed
		} else {
			s.writeError(w, http.StatusBadRequest, "Invalid sold_date format. Use RFC3339 format")
			return
		}
	}

	// Validation: If status is SOLD, require customer details
	if req.SaleStatus == "SOLD" {
		if req.SoldToName == nil || *req.SoldToName == "" {
			s.writeError(w, http.StatusBadRequest, "Customer name is required when status is SOLD")
			return
		}
		if req.Revenue == nil || *req.Revenue <= 0 {
			s.writeError(w, http.StatusBadRequest, "Revenue is required when status is SOLD")
			return
		}
		if soldDate == nil {
			now := time.Now()
			soldDate = &now
		}
	}

	err = s.vehicleService.UpdateSalesDetails(id, req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Sales details updated successfully"})
}

func (s *APIServer) updateVehicle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var req request.UpdateVehicleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate condition status if provided
	if req.ConditionStatus != nil {
		validConditions := map[string]bool{"REGISTERED": true, "UNREGISTERED": true}
		if !validConditions[*req.ConditionStatus] {
			s.writeError(w, http.StatusBadRequest, "Invalid condition status")
			return
		}
	}

	// Validate currency if provided
	if req.Currency != nil && *req.Currency != "" {
		validCurrencies := map[string]bool{"JPY": true, "USD": true, "LKR": true}
		if !validCurrencies[*req.Currency] {
			s.writeError(w, http.StatusBadRequest, "Invalid currency")
			return
		}
	}

	err = s.vehicleService.UpdateVehicleDetails(id, &req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle details updated successfully"})
}

func (s *APIServer) createVehicleMake(w http.ResponseWriter, r *http.Request) {
	var req request.CreateVehicleMake

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if *req.MakeName == "" {
		s.writeError(w, http.StatusBadRequest, "Make name is required")
		return
	}

	// Set defaults
	if req.CountryOrigin == nil || *req.CountryOrigin == "" {
		defaultVal := "Japan"
		req.CountryOrigin = &defaultVal
	}
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	make, err := s.vehicleService.CreateVehicleMake(req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			s.writeError(w, http.StatusConflict, "Make name already exists")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    make,
		"message": "Vehicle make created successfully",
	})
}

func (s *APIServer) getVehicleMakes(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	makes, err := s.vehicleService.GetAllVehicleMakes(activeOnly)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": makes,
		"meta": map[string]interface{}{
			"total":       len(makes),
			"active_only": activeOnly,
		},
	})
}

func (s *APIServer) updateVehicleMake(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid make ID")
		return
	}

	var req request.CreateVehicleMake

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = s.vehicleService.UpdateVehicleMake(id, req.MakeName, req.CountryOrigin, req.IsActive)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle make updated successfully"})
}

func (s *APIServer) createVehicleModel(w http.ResponseWriter, r *http.Request) {
	var req request.CreateVehicleModel

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.MakeID <= 0 {
		s.writeError(w, http.StatusBadRequest, "Valid make ID is required")
		return
	}
	if *req.ModelName == "" {
		s.writeError(w, http.StatusBadRequest, "Model name is required")
		return
	}

	// Set defaults
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	model, err := s.vehicleService.CreateVehicleModel(req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			s.writeError(w, http.StatusConflict, "Model name already exists for this make")
			return
		}
		if strings.Contains(err.Error(), "foreign key") {
			s.writeError(w, http.StatusBadRequest, "Invalid make ID")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    model,
		"message": "Vehicle model created successfully",
	})
}

func (s *APIServer) getVehicleModels(w http.ResponseWriter, r *http.Request) {
	makeIDStr := r.URL.Query().Get("make_id")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var makeID *int
	if makeIDStr != "" {
		if id, err := strconv.Atoi(makeIDStr); err == nil {
			makeID = &id
		} else {
			s.writeError(w, http.StatusBadRequest, "Invalid make_id parameter")
			return
		}
	}

	models, err := s.vehicleService.GetVehicleModels(makeID, activeOnly)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": models,
		"meta": map[string]interface{}{
			"total":       len(models),
			"make_id":     makeID,
			"active_only": activeOnly,
		},
	})
}

func (s *APIServer) getVehicleModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid model ID")
		return
	}

	model, err := s.vehicleService.GetVehicleModelByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeError(w, http.StatusNotFound, "Vehicle model not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": model})
}

func (s *APIServer) updateVehicleModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid model ID")
		return
	}

	var req request.CreateVehicleModel

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = s.vehicleService.UpdateVehicleModel(id, req.ModelName, req.BodyType,
		req.FuelType, req.TransmissionType, req.EngineSizeCC, req.IsActive)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle model updated successfully"})
}

// Order handlers
func (s *APIServer) getOrders(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	orders, err := s.orderService.GetAllOrders(limit, offset)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data": orders,
		"meta": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": len(orders),
		},
	}
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) getOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := s.orderService.GetOrderByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeError(w, http.StatusNotFound, "Order not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": order})
}

func (s *APIServer) createOrder(w http.ResponseWriter, r *http.Request) {
	var req request.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.CustomerName == "" || req.ContactNumber == "" || req.PreferredMake == "" || req.PreferredModel == "" {
		s.writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	order, err := s.orderService.CreateOrder(req)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data":    order,
		"message": "Order created successfully",
	}
	s.writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Status == "" {
		s.writeError(w, http.StatusBadRequest, "Status is required")
		return
	}

	err = s.orderService.UpdateOrderStatus(id, req.Status)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"message": "Order status updated successfully"})
}

// Analytics handlers
func (s *APIServer) getDashboardStats(w http.ResponseWriter, r *http.Request) {
	// This would typically aggregate data from multiple tables
	stats := map[string]interface{}{
		"total_vehicles":    150,
		"vehicles_in_stock": 45,
		"vehicles_sold":     105,
		"pending_orders":    8,
		"total_revenue":     450000000, // LKR
		"total_profit":      75000000,  // LKR
		"profit_margin":     16.67,     // %
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": stats})
}

func (s *APIServer) getSalesSummary(w http.ResponseWriter, r *http.Request) {
	// Monthly sales summary - would typically come from database
	summary := []map[string]interface{}{
		{
			"month":         "2024-09",
			"vehicles_sold": 12,
			"total_revenue": 45000000,
			"total_profit":  7500000,
			"profit_margin": 16.67,
		},
		{
			"month":         "2024-08",
			"vehicles_sold": 15,
			"total_revenue": 52000000,
			"total_profit":  8200000,
			"profit_margin": 15.77,
		},
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": summary})
}

func (s *APIServer) getInventoryStatus(w http.ResponseWriter, r *http.Request) {
	// Inventory breakdown by status
	inventory := []map[string]interface{}{
		{
			"shipping_status": "DELIVERED",
			"sale_status":     "SOLD",
			"count":           105,
			"investment":      315000000,
		},
		{
			"shipping_status": "CLEARED",
			"sale_status":     "AVAILABLE",
			"count":           25,
			"investment":      75000000,
		},
		{
			"shipping_status": "ARRIVED",
			"sale_status":     "AVAILABLE",
			"count":           15,
			"investment":      45000000,
		},
		{
			"shipping_status": "SHIPPED",
			"sale_status":     "AVAILABLE",
			"count":           5,
			"investment":      15000000,
		},
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": inventory})
}

func (s *APIServer) Start(port string) error {
	log.Printf("Starting server on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}
