package controllers

import (
	"car_service/dto/request"
	"car_service/internal/constants"
	"car_service/middleware"
	"car_service/services"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type VehicleShareController struct {
	vehicleService *services.VehicleService
	s3Service      *services.S3Service
	router         *mux.Router
	introspectURL  string
}

func NewVehicleShareController(vehicleService *services.VehicleService, s3Service *services.S3Service, router *mux.Router, introspectURL string) *VehicleShareController {
	return &VehicleShareController{
		vehicleService: vehicleService,
		s3Service:      s3Service,
		router:         router,
		introspectURL:  introspectURL,
	}
}

func (vc *VehicleShareController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (vc *VehicleShareController) writeError(w http.ResponseWriter, status int, message string) {
	vc.writeJSON(w, status, map[string]string{"error": message})
}

func (vc *VehicleShareController) SetupRoutes() {

	api := vc.router.PathPrefix("/car-service/api/v1").Subrouter()
	authMiddleware := middleware.NewAuthMiddleware(vc.introspectURL)

	// Vehicle share routes
	share := api.PathPrefix("/share").Subrouter()

	// Protected endpoint: Generate share token for a vehicle
	share.Handle("/vehicle/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.generateVehicleShareToken), constants.VEHICLE_ACCESS)).Methods("POST")

	// Public endpoint: Get vehicle data using share token (no authentication required)
	share.Handle("/vehicle/public/{shareToken}", http.HandlerFunc(vc.getPublicShareVehicleData)).Methods("GET")

}

func (vc *VehicleShareController) generateVehicleShareToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var req request.PublicTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate expire_in_days
	if req.ExpireInDays <= 0 {
		req.ExpireInDays = 7 // Default to 7 days
	}
	if req.ExpireInDays > 365 {
		vc.writeError(w, http.StatusBadRequest, "Expiration period cannot exceed 365 days")
		return
	}

	// Validate include_details
	validDetails := map[string]bool{
		"shipping":  true,
		"financial": true,
		"purchase":  true,
		"images":    true,
	}
	for _, detail := range req.IncludeDetails {
		if !validDetails[detail] {
			vc.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid detail type: %s", detail))
			return
		}
	}

	shareToken, err := vc.vehicleService.GenerateShareToken(r.Context(), vehicleID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			vc.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		vc.writeError(w, http.StatusInternalServerError, "Failed to generate share token")
		return
	}

	vc.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"token":      shareToken.Token,
		"vehicle_id": shareToken.VehicleID,
		"expires_at": shareToken.ExpiresAt,
		"share_url":  fmt.Sprintf("/car-service/api/v1/share/vehicle/public/%s", shareToken.Token),
	})
}

func (vc *VehicleShareController) getPublicShareVehicleData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shareToken := vars["shareToken"]

	if shareToken == "" {
		vc.writeError(w, http.StatusBadRequest, "Share token is required")
		return
	}

	publicData, err := vc.vehicleService.GetPublicVehicleData(r.Context(), shareToken)
	if err != nil {
		if err.Error() == "invalid or expired share token" {
			vc.writeError(w, http.StatusNotFound, "Invalid or expired share token")
			return
		}
		vc.writeError(w, http.StatusInternalServerError, "Failed to retrieve vehicle data")
		return
	}

	vc.writeJSON(w, http.StatusOK, publicData)
}
