package controllers

import (
	"car_service/dto/request"
	"car_service/internal/constants"
	"car_service/logger"
	"car_service/middleware"
	"car_service/services"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type CustomerController struct {
	customerService *services.CustomerService
	router          *mux.Router
	introspectURL   string
}

func NewCustomerController(router *mux.Router, introspectURL string, customerService *services.CustomerService) *CustomerController {
	return &CustomerController{
		customerService: customerService,
		router:          router,
		introspectURL:   introspectURL,
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
	logger.WithFields(map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Create customer request received")

	var req request.CreateCustomerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithField("error", err.Error()).Warn("Invalid JSON in create customer request")
		cc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Get authorization header for notification service
	authHeader := r.Header.Get("Authorization")

	customer, err := cc.customerService.CreateCustomer(r.Context(), req, authHeader)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			cc.writeError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			cc.writeError(w, http.StatusBadRequest, err.Error())
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
	searchTerm := r.URL.Query().Get("q")

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10 // Default limit
	}
	offset := (page - 1) * limit

	var customerTypePtr *string
	if customerType != "" {
		customerTypePtr = &customerType
	}

	customers, total, err := cc.customerService.GetAllCustomers(r.Context(), limit, offset, customerTypePtr, activeOnly, searchTerm)
	if err != nil {
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	meta := map[string]interface{}{
		"total":         total,
		"count":         len(customers),
		"page":          page,
		"limit":         limit,
		"customer_type": customerType,
		"active_only":   activeOnly,
	}

	// Add search_term to meta if it was provided
	if searchTerm != "" {
		meta["search_term"] = searchTerm
	}

	cc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": customers,
		"meta": meta,
	})
}

func (cc *CustomerController) getCustomerByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		cc.writeError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	customer, err := cc.customerService.GetCustomerByID(r.Context(), id)
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

	err = cc.customerService.UpdateCustomer(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			cc.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
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

	// Extract authorization header from request
	authHeader := r.Header.Get("Authorization")

	err = cc.customerService.DeleteCustomer(r.Context(), id, authHeader)
	if err != nil {
		cc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cc.writeJSON(w, http.StatusOK, map[string]string{"message": "Customer deleted successfully"})
}

func (cc *CustomerController) searchCustomers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	searchTerm := r.URL.Query().Get("q")

	customers, err := cc.customerService.SearchCustomers(r.Context(), searchTerm)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			cc.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
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
