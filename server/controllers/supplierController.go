package controllers

import (
	"car_service/dto/request"
	"car_service/internal/constants"
	"car_service/middleware"
	"car_service/services"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type SupplierController struct {
	supplierService *services.SupplierService
	router          *mux.Router
	introspectURL   string
}

func NewSupplierController(router *mux.Router, introspectURL string, supplierService *services.SupplierService) *SupplierController {
	return &SupplierController{
		supplierService: supplierService,
		router:          router,
		introspectURL:   introspectURL,
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
	suppliers.Handle("/{id:[0-9]+}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.getSupplierByID(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// POST create supplier
	suppliers.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.createSupplier(w, r, db)
	}), constants.VEHICLE_CREATE)).Methods("POST")

	// PUT update supplier
	suppliers.Handle("/{id:[0-9]+}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc.updateSupplier(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("PUT")

	// DELETE supplier (soft delete)
	suppliers.Handle("/{id:[0-9]+}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Get authorization header for notification service
	authHeader := r.Header.Get("Authorization")

	supplier, err := sc.supplierService.CreateSupplier(r.Context(), req, authHeader)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			sc.writeError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			sc.writeError(w, http.StatusBadRequest, err.Error())
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

	var supplierTypePtr *string
	if supplierType != "" {
		supplierTypePtr = &supplierType
	}

	suppliers, total, err := sc.supplierService.GetAllSuppliers(r.Context(), limit, offset, supplierTypePtr, activeOnly, searchTerm)
	if err != nil {
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	meta := map[string]interface{}{
		"total":         total,
		"count":         len(suppliers),
		"page":          page,
		"limit":         limit,
		"supplier_type": supplierType,
		"active_only":   activeOnly,
	}

	// Add search_term to meta if it was provided
	if searchTerm != "" {
		meta["search_term"] = searchTerm
	}

	sc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": suppliers,
		"meta": meta,
	})
}

func (sc *SupplierController) getSupplierByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		sc.writeError(w, http.StatusBadRequest, "Invalid supplier ID")
		return
	}

	supplier, err := sc.supplierService.GetSupplierByID(r.Context(), id)
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

	err = sc.supplierService.UpdateSupplier(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			sc.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
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

	err = sc.supplierService.DeleteSupplier(r.Context(), id)
	if err != nil {
		sc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sc.writeJSON(w, http.StatusOK, map[string]string{"message": "Supplier deleted successfully"})
}

func (sc *SupplierController) searchSuppliers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	searchTerm := r.URL.Query().Get("q")

	suppliers, err := sc.supplierService.SearchSuppliers(r.Context(), searchTerm)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			sc.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
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
