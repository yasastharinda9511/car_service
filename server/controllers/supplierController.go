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

type SupplierController struct {
	supplierRepository *repository.SupplierRepository
	router             *mux.Router
	introspectURL      string
}

func NewSupplierController(router *mux.Router, introspectURL string) *SupplierController {
	return &SupplierController{
		supplierRepository: repository.NewSupplierRepository(),
		router:             router,
		introspectURL:      introspectURL,
	}
}

func (sc *SupplierController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (sc *SupplierController) writeError(w http.ResponseWriter, status int, message string) {
	sc.writeJSON(w, status, map[string]string{"error": message})
}

func (sc *SupplierController) SetupRoutes(db *sql.DB) {
	api := sc.router.PathPrefix("/car-service/api/v1").Subrouter()
	authMiddleware := middleware.NewAuthMiddleware(sc.introspectURL)

	// Supplier routes
	suppliers := api.PathPrefix("/suppliers").Subrouter()

	// GET all suppliers
	suppliers.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.getSuppliers(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// GET supplier by ID
	suppliers.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.getSupplierByID(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// POST create supplier
	suppliers.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.createSupplier(w, r, db)
	}), constants.VEHICLE_CREATE)).Methods("POST")

	// PUT update supplier
	suppliers.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.updateSupplier(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("PUT")

	// DELETE supplier (soft delete)
	suppliers.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.deleteSupplier(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("DELETE")

	// GET search suppliers
	suppliers.Handle("/search", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.searchSuppliers(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")
}

func (sc *SupplierController) createSupplier(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req request.CreateSupplierRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.SupplierName == "" {
		sc.writeError(w, http.StatusBadRequest, "Supplier name is required")
		return
	}

	if req.SupplierType == "" {
		sc.writeError(w, http.StatusBadRequest, "Supplier type is required")
		return
	}

	// Validate supplier type enum
	validTypes := map[string]bool{"AUCTION": true, "DEALER": true, "INDIVIDUAL": true}
	if !validTypes[req.SupplierType] {
		sc.writeError(w, http.StatusBadRequest, "Invalid supplier type. Must be AUCTION, DEALER, or INDIVIDUAL")
		return
	}

	// Set default for is_active
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	supplier, err := sc.supplierRepository.CreateSupplier(r.Context(), db, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			sc.writeError(w, http.StatusConflict, "Supplier with this information already exists")
			return
		}
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    supplier,
		"message": "Supplier created successfully",
	})
}

func (sc *SupplierController) getSuppliers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	supplierType := r.URL.Query().Get("supplier_type")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var supplierTypePtr *string
	if supplierType != "" {
		supplierTypePtr = &supplierType
	}

	suppliers, err := sc.supplierRepository.GetAllSuppliers(r.Context(), db, supplierTypePtr, activeOnly)
	if err != nil {
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": suppliers,
		"meta": map[string]interface{}{
			"total":         len(suppliers),
			"supplier_type": supplierType,
			"active_only":   activeOnly,
		},
	})
}

func (sc *SupplierController) getSupplierByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		sc.writeError(w, http.StatusBadRequest, "Invalid supplier ID")
		return
	}

	supplier, err := sc.supplierRepository.GetSupplierByID(r.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			sc.writeError(w, http.StatusNotFound, "Supplier not found")
			return
		}
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusOK, map[string]interface{}{"data": supplier})
}

func (sc *SupplierController) updateSupplier(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		sc.writeError(w, http.StatusBadRequest, "Invalid supplier ID")
		return
	}

	var req request.UpdateSupplierRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate supplier type if provided
	if req.SupplierType != nil && *req.SupplierType != "" {
		validTypes := map[string]bool{"AUCTION": true, "DEALER": true, "INDIVIDUAL": true}
		if !validTypes[*req.SupplierType] {
			sc.writeError(w, http.StatusBadRequest, "Invalid supplier type. Must be AUCTION, DEALER, or INDIVIDUAL")
			return
		}
	}

	err = sc.supplierRepository.UpdateSupplier(r.Context(), db, id, req)
	if err != nil {
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusOK, map[string]string{"message": "Supplier updated successfully"})
}

func (sc *SupplierController) deleteSupplier(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		sc.writeError(w, http.StatusBadRequest, "Invalid supplier ID")
		return
	}

	err = sc.supplierRepository.DeleteSupplier(r.Context(), db, id)
	if err != nil {
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusOK, map[string]string{"message": "Supplier deleted successfully"})
}

func (sc *SupplierController) searchSuppliers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		sc.writeError(w, http.StatusBadRequest, "Search term 'q' is required")
		return
	}

	suppliers, err := sc.supplierRepository.SearchSuppliers(r.Context(), db, searchTerm)
	if err != nil {
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": suppliers,
		"meta": map[string]interface{}{
			"search_term": searchTerm,
			"total":       len(suppliers),
		},
	})
}
