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

type VehicleMakeController struct {
	makeRepository *repository.VehicleMakeRepository
	router         *mux.Router
	introspectURL  string
}

func NewVehicleMakeController(router *mux.Router, introspectURL string) *VehicleMakeController {
	return &VehicleMakeController{
		makeRepository: repository.NewVehicleMakeRepository(),
		router:         router,
		introspectURL:  introspectURL,
	}
}

func (mc *VehicleMakeController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (mc *VehicleMakeController) writeError(w http.ResponseWriter, status int, message string) {
	mc.writeJSON(w, status, map[string]string{"error": message})
}

func (mc *VehicleMakeController) SetupRoutes(db *sql.DB) {
	api := mc.router.PathPrefix("/car-service/api/v1").Subrouter()
	authMiddleware := middleware.NewAuthMiddleware(mc.introspectURL)

	// Vehicle makes routes
	makes := api.PathPrefix("/makes").Subrouter()

	// GET all makes
	makes.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.getVehicleMakes(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// POST create new make
	makes.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.createVehicleMake(w, r, db)
	}), constants.VEHICLE_CREATE)).Methods("POST")

	// PUT update make
	makes.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.updateVehicleMake(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("PUT")
}

func (mc *VehicleMakeController) createVehicleMake(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req request.CreateVehicleMake

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.MakeName == nil || *req.MakeName == "" {
		mc.writeError(w, http.StatusBadRequest, "Make name is required")
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

	make, err := mc.makeRepository.CreateVehicleMake(r.Context(), db, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			mc.writeError(w, http.StatusConflict, "Make name already exists")
			return
		}
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    make,
		"message": "Vehicle make created successfully",
	})
}

func (mc *VehicleMakeController) getVehicleMakes(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	makes, err := mc.makeRepository.GetAllVehicleMakes(r.Context(), db, activeOnly)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": makes,
		"meta": map[string]interface{}{
			"total":       len(makes),
			"active_only": activeOnly,
		},
	})
}

func (mc *VehicleMakeController) updateVehicleMake(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid make ID")
		return
	}

	var req request.CreateVehicleMake

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = mc.makeRepository.UpdateVehicleMake(r.Context(), db, id, req.MakeName, req.CountryOrigin, req.IsActive)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle make updated successfully"})
}
