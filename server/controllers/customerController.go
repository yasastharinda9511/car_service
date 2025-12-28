package controllers

import (
	"car_service/dto/request"
	"car_service/internal/constants"
	"car_service/middleware"
	"car_service/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type CustomerController struct {
	customerRepository *repository.CustomerRepository
	router             *mux.Router
	introspectURL      string
}

func NewCustomerController(router *mux.Router, introspectURL string) *CustomerController {
	return &CustomerController{
		customerRepository: repository.NewCustomerRepository(),
		router:             router,
		introspectURL:      introspectURL,
	}
}

func (cc *CustomerController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (cc *CustomerController) writeError(w http.ResponseWriter, status int, message string) {
	cc.writeJSON(w, status, map[string]string{"error": message})
}

func (cc *CustomerController) SetupRoutes(db *sql.DB) {
	api := cc.router.PathPrefix("/car-service/api/v1").Subrouter()
	authMiddleware := middleware.NewAuthMiddleware(cc.introspectURL)

	// Customer routes
	customers := api.PathPrefix("/customers").Subrouter()

	// GET all customers
	customers.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc.getCustomers(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// POST create customer
	customers.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc.createCustomer(w, r, db)
	}), constants.VEHICLE_CREATE)).Methods("POST")

	// GET search customers
	customers.Handle("/search", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc.searchCustomers(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// GET customer by ID
	customers.Handle("/{id:[0-9]+}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc.getCustomerByID(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// PUT update customer
	customers.Handle("/{id:[0-9]+}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc.updateCustomer(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("PUT")

	// DELETE customer (soft delete)
	customers.Handle("/{id}:[0-9]+", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc.deleteCustomer(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("DELETE")
}

func (cc *CustomerController) createCustomer(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req request.CreateCustomerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.CustomerName == "" {
		cc.writeError(w, http.StatusBadRequest, "Customer name is required")
		return
	}

	if req.CustomerType == "" {
		cc.writeError(w, http.StatusBadRequest, "Customer type is required")
		return
	}

	// Validate customer type enum
	validTypes := map[string]bool{"INDIVIDUAL": true, "BUSINESS": true}
	if !validTypes[req.CustomerType] {
		cc.writeError(w, http.StatusBadRequest, "Invalid customer type. Must be INDIVIDUAL or BUSINESS")
		return
	}

	// Set default for is_active
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	customer, err := cc.customerRepository.CreateCustomer(r.Context(), db, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			cc.writeError(w, http.StatusConflict, "Customer with this information already exists")
			return
		}
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    customer,
		"message": "Customer created successfully",
	})
}

func (cc *CustomerController) getCustomers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	customerType := r.URL.Query().Get("customer_type")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var customerTypePtr *string
	if customerType != "" {
		customerTypePtr = &customerType
	}

	customers, err := cc.customerRepository.GetAllCustomers(r.Context(), db, customerTypePtr, activeOnly)
	if err != nil {
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": customers,
		"meta": map[string]interface{}{
			"total":         len(customers),
			"customer_type": customerType,
			"active_only":   activeOnly,
		},
	})
}

func (cc *CustomerController) getCustomerByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		cc.writeError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	customer, err := cc.customerRepository.GetCustomerByID(r.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			cc.writeError(w, http.StatusNotFound, "Customer not found")
			return
		}
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusOK, map[string]interface{}{"data": customer})
}

func (cc *CustomerController) updateCustomer(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		cc.writeError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var req request.UpdateCustomerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate customer type if provided
	if req.CustomerType != nil && *req.CustomerType != "" {
		validTypes := map[string]bool{"INDIVIDUAL": true, "BUSINESS": true}
		if !validTypes[*req.CustomerType] {
			cc.writeError(w, http.StatusBadRequest, "Invalid customer type. Must be INDIVIDUAL or BUSINESS")
			return
		}
	}

	err = cc.customerRepository.UpdateCustomer(r.Context(), db, id, req)
	if err != nil {
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusOK, map[string]string{"message": "Customer updated successfully"})
}

func (cc *CustomerController) deleteCustomer(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		cc.writeError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	err = cc.customerRepository.DeleteCustomer(r.Context(), db, id)
	if err != nil {
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusOK, map[string]string{"message": "Customer deleted successfully"})
}

func (cc *CustomerController) searchCustomers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		cc.writeError(w, http.StatusBadRequest, "Search term 'q' is required")
		return
	}

	customers, err := cc.customerRepository.SearchCustomers(r.Context(), db, searchTerm)
	if err != nil {
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": customers,
		"meta": map[string]interface{}{
			"search_term": searchTerm,
			"total":       len(customers),
		},
	})
}
