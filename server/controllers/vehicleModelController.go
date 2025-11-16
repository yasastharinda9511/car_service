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

type VehicleModelController struct {
	modelRepository *repository.VehicleModelRepository
	router          *mux.Router
	introspectURL   string
}

func NewVehicleModelController(router *mux.Router, introspectURL string) *VehicleModelController {
	return &VehicleModelController{
		modelRepository: repository.NewVehicleModelRepository(),
		router:          router,
		introspectURL:   introspectURL,
	}
}

func (mc *VehicleModelController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (mc *VehicleModelController) writeError(w http.ResponseWriter, status int, message string) {
	mc.writeJSON(w, status, map[string]string{"error": message})
}

func (mc *VehicleModelController) SetupRoutes(db *sql.DB) {
	api := mc.router.PathPrefix("/car-service/api/v1").Subrouter()
	authMiddleware := middleware.NewAuthMiddleware(mc.introspectURL)

	// Vehicle models routes
	models := api.PathPrefix("/models").Subrouter()

	// GET all models
	models.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.getVehicleModels(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// GET model by ID
	models.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.getVehicleModel(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")

	// POST create new model
	models.Handle("", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.createVehicleModel(w, r, db)
	}), constants.VEHICLE_CREATE)).Methods("POST")

	// PUT update model
	models.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.updateVehicleModel(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("PUT")
}

func (mc *VehicleModelController) createVehicleModel(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req request.CreateVehicleModel

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.MakeID <= 0 {
		mc.writeError(w, http.StatusBadRequest, "Valid make ID is required")
		return
	}
	if req.ModelName == nil || *req.ModelName == "" {
		mc.writeError(w, http.StatusBadRequest, "Model name is required")
		return
	}

	// Set defaults
	if req.IsActive == nil {
		defaultActive := true
		req.IsActive = &defaultActive
	}

	model, err := mc.modelRepository.CreateVehicleModel(r.Context(), db, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			mc.writeError(w, http.StatusConflict, "Model name already exists for this make")
			return
		}
		if strings.Contains(err.Error(), "foreign key") {
			mc.writeError(w, http.StatusBadRequest, "Invalid make ID")
			return
		}
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data":    model,
		"message": "Vehicle model created successfully",
	})
}

func (mc *VehicleModelController) getVehicleModels(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	makeIDStr := r.URL.Query().Get("make_id")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var makeID *int
	if makeIDStr != "" {
		if id, err := strconv.Atoi(makeIDStr); err == nil {
			makeID = &id
		} else {
			mc.writeError(w, http.StatusBadRequest, "Invalid make_id parameter")
			return
		}
	}

	models, err := mc.modelRepository.GetVehicleModels(r.Context(), db, makeID, activeOnly)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": models,
		"meta": map[string]interface{}{
			"total":       len(models),
			"make_id":     makeID,
			"active_only": activeOnly,
		},
	})
}

func (mc *VehicleModelController) getVehicleModel(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid model ID")
		return
	}

	model, err := mc.modelRepository.GetVehicleModelByID(r.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			mc.writeError(w, http.StatusNotFound, "Vehicle model not found")
			return
		}
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusOK, map[string]interface{}{"data": model})
}

func (mc *VehicleModelController) updateVehicleModel(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid model ID")
		return
	}

	var req request.CreateVehicleModel

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = mc.modelRepository.UpdateVehicleModel(r.Context(), db, id, req.ModelName, req.BodyType,
		req.FuelType, req.TransmissionType, req.EngineSizeCC, req.IsActive)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mc.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle model updated successfully"})
}
