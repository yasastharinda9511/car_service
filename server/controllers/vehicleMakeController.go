package controllers

import (
	"car_service/dto/request"
	"car_service/internal/constants"
	"car_service/middleware"
	"car_service/repository"
	"car_service/services"
	"car_service/util"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type VehicleMakeController struct {
	makeRepository *repository.VehicleMakeRepository
	s3Service      *services.S3Service
	router         *mux.Router
	introspectURL  string
}

func NewVehicleMakeController(router *mux.Router, introspectURL string, s3Service *services.S3Service) *VehicleMakeController {
	return &VehicleMakeController{
		makeRepository: repository.NewVehicleMakeRepository(),
		s3Service:      s3Service,
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

	// POST upload make logo
	makes.Handle("/{id}/logo", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.uploadMakeLogo(w, r, db)
	}), constants.VEHICLE_EDIT)).Methods("POST")

	// GET make logo (presigned URL)
	makes.Handle("/{id}/logo", authMiddleware.Authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mc.getMakeLogo(w, r, db)
	}), constants.VEHICLE_ACCESS)).Methods("GET")
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

func (mc *VehicleMakeController) uploadMakeLogo(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid make ID")
		return
	}

	// Check if make exists
	make, err := mc.makeRepository.GetVehicleMakeByID(r.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			mc.writeError(w, http.StatusNotFound, "Vehicle make not found")
			return
		}
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(5 << 20) // 5MB limit for logos
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Unable to parse form")
		return
	}

	// Get the logo file
	file, fileHeader, err := r.FormFile("logo")
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Logo file is required")
		return
	}
	defer file.Close()

	// Validate file type (images only)
	contentType := fileHeader.Header.Get("Content-Type")
	if !util.IsValidImageType(contentType) {
		mc.writeError(w, http.StatusBadRequest, "Invalid file type. Only JPEG, PNG, GIF allowed")
		return
	}

	// Validate file size (max 5MB)
	if fileHeader.Size > 5<<20 {
		mc.writeError(w, http.StatusBadRequest, "File too large. Maximum size is 5MB")
		return
	}

	// Upload to S3
	pathPrefix := fmt.Sprintf("makes/%d", id)
	if mc.s3Service == nil {
		mc.writeError(w, http.StatusInternalServerError, "S3 service not configured")
		return
	}

	result, err := mc.s3Service.UploadFile(r.Context(), file, fileHeader, pathPrefix)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload logo: %v", err))
		return
	}

	// Update logo_url in database
	err = mc.makeRepository.UpdateMakeLogo(r.Context(), db, id, result.Key)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, "Failed to update logo URL in database")
		return
	}

	// Generate presigned URL for response
	presignedURL, err := mc.s3Service.GetPresignedURL(r.Context(), result.Key, 15)
	if err != nil {
		mc.writeError(w, http.StatusInternalServerError, "Failed to generate presigned URL")
		return
	}

	mc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Logo uploaded successfully",
		"data": map[string]interface{}{
			"make_id":       make.ID,
			"make_name":     make.MakeName,
			"logo_url":      result.Key,
			"presigned_url": presignedURL,
			"file_size":     fileHeader.Size,
		},
	})
}

func (mc *VehicleMakeController) getMakeLogo(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		mc.writeError(w, http.StatusBadRequest, "Invalid make ID")
		return
	}

	// Get make details
	make, err := mc.makeRepository.GetVehicleMakeByID(r.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			mc.writeError(w, http.StatusNotFound, "Vehicle make not found")
			return
		}
		mc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if logo exists
	if make.LogoURL == nil || *make.LogoURL == "" {
		mc.writeError(w, http.StatusNotFound, "No logo uploaded for this make")
		return
	}

	// Check if file exists in S3
	if mc.s3Service != nil {
		exists, err := mc.s3Service.CheckIfFileExists(r.Context(), *make.LogoURL)
		if err != nil || !exists {
			mc.writeError(w, http.StatusNotFound, "Logo not found in storage")
			return
		}

		// Generate presigned URL (valid for 15 minutes)
		presignedURL, err := mc.s3Service.GetPresignedURL(r.Context(), *make.LogoURL, 15)
		if err != nil {
			mc.writeError(w, http.StatusInternalServerError, "Failed to generate logo URL")
			return
		}

		mc.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": presignedURL,
			"metadata": map[string]interface{}{
				"make_id":   make.ID,
				"make_name": make.MakeName,
				"logo_url":  *make.LogoURL,
			},
		})
	}
}
