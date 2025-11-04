package controllers

import (
	"car_service/dto/request"
	"car_service/filters"
	"car_service/services"
	"net/http"

	"car_service/entity"
	"car_service/util"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type VehicleController struct {
	vehicleService *services.VehicleService
	router         *mux.Router
}

func NewVehicleController(vehicleService *services.VehicleService, router *mux.Router) *VehicleController {
	return &VehicleController{
		vehicleService: vehicleService,
		router:         router,
	}
}

func (vc *VehicleController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (vc *VehicleController) writeError(w http.ResponseWriter, status int, message string) {
	vc.writeJSON(w, status, map[string]string{"error": message})
}

func (vc *VehicleController) SetupRoutes() {

	api := vc.router.PathPrefix("/car-service/api/v1").Subrouter()

	// Vehicle routes
	vehicles := api.PathPrefix("/vehicles").Subrouter()
	vehicles.HandleFunc("", vc.getVehicles).Methods("GET")
	vehicles.HandleFunc("/{id}", vc.getVehicle).Methods("GET")
	vehicles.HandleFunc("", vc.createVehicle).Methods("POST")
	vehicles.HandleFunc("/upload-image/{filename}", vc.serveImageHandler).Methods("GET")
	vehicles.HandleFunc("/upload-image/{id}", vc.uploadImageHandler).Methods("POST")

	vehicles.HandleFunc("/{id}/shipping", vc.updateShipping).Methods("PUT")
	vehicles.HandleFunc("/{id}/purchase", vc.updatePurchase).Methods("PUT")
	vehicles.HandleFunc("/{id}/financials", vc.updateFinancials).Methods("PUT")
	vehicles.HandleFunc("/{id}/sales", vc.updateSales).Methods("PUT")
	vehicles.HandleFunc("/{id}", vc.updateVehicle).Methods("PUT")

}

func (vc *VehicleController) getVehicles(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	vehicleFilter := filters.NewVehicleFilters()
	vehicleFilter.GetValuesFromRequest(r)

	vehicles, err := vc.vehicleService.GetAllVehicles(r.Context(), limit, offset, vehicleFilter)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vehicles.Meta.Page = page
	vc.writeJSON(w, http.StatusOK, vehicles)
}

func (vc *VehicleController) getVehicle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	vehicle, err := vc.vehicleService.GetVehicleByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			vc.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{"data": vehicle})
}

func (vc *VehicleController) createVehicle(w http.ResponseWriter, r *http.Request) {
	var req request.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validation
	if req.Make == "" || req.Model == "" || req.ChassisID == "" {
		vc.writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	vehicle, err := vc.vehicleService.CreateVehicle(r.Context(), req)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusCreated, map[string]interface{}{"data": vehicle})
}

func (vc *VehicleController) updateShipping(w http.ResponseWriter, r *http.Request) {
	var req request.ShippingDetailsRequest

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err = vc.vehicleService.UpdateShippingStatus(r.Context(), id, req)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{"message": "Shipping status updated successfully"})
}

func (vc *VehicleController) updatePurchase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req request.PurchaseRequest
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.PurchaseDate != nil {
		// Convert back to string for validation (if necessary)
		parsed, err := time.Parse(time.RFC3339, req.PurchaseDate.Format(time.RFC3339))
		if err != nil {
			vc.writeError(w, http.StatusBadRequest, "Invalid purchase_date format. Use RFC3339 format")
			return
		}
		req.PurchaseDate = &parsed
	}

	err = vc.vehicleService.UpdatePurchaseDetails(r.Context(), id, &req)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{"message": "Purchase details updated successfully"})
}

func (vc *VehicleController) updateFinancials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var req request.FinancialDetailsRequest
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required field
	if req.TotalCostLKR <= 0 {
		vc.writeError(w, http.StatusBadRequest, "Total cost must be greater than 0")
		return
	}

	err = vc.vehicleService.UpdateFinancialDetails(r.Context(), id, &req)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{"message": "Financial details updated successfully"})
}

func (vc *VehicleController) updateSales(w http.ResponseWriter, r *http.Request) {
	var req request.SalesDetailsRequest
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required field
	if req.SaleStatus == "" {
		vc.writeError(w, http.StatusBadRequest, "Sale status is required")
		return
	}

	// Validate sale status enum values
	validStatuses := map[string]bool{
		"AVAILABLE": true, "RESERVED": true, "SOLD": true, "CANCELLED": true,
	}
	if !validStatuses[req.SaleStatus] {
		vc.writeError(w, http.StatusBadRequest, "Invalid sale status")
		return
	}

	// Parse sold date if provided
	var soldDate *time.Time
	if req.SoldDate != nil && *req.SoldDate != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.SoldDate); err == nil {
			soldDate = &parsed
		} else {
			vc.writeError(w, http.StatusBadRequest, "Invalid sold_date format. Use RFC3339 format")
			return
		}
	}

	// Validation: If status is SOLD, require customer details
	if req.SaleStatus == "SOLD" {
		if req.SoldToName == nil || *req.SoldToName == "" {
			vc.writeError(w, http.StatusBadRequest, "Customer name is required when status is SOLD")
			return
		}
		if req.Revenue == nil || *req.Revenue <= 0 {
			vc.writeError(w, http.StatusBadRequest, "Revenue is required when status is SOLD")
			return
		}
		if soldDate == nil {
			now := time.Now()
			soldDate = &now
		}
	}

	err = vc.vehicleService.UpdateSalesDetails(r.Context(), id, &req)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{"message": "Sales details updated successfully"})
}

func (vc *VehicleController) updateVehicle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var req request.UpdateVehicleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate condition status if provided
	if req.ConditionStatus != nil {
		validConditions := map[string]bool{"REGISTERED": true, "UNREGISTERED": true}
		if !validConditions[*req.ConditionStatus] {
			vc.writeError(w, http.StatusBadRequest, "Invalid condition status")
			return
		}
	}

	// Validate currency if provided
	if req.Currency != nil && *req.Currency != "" {
		validCurrencies := map[string]bool{"JPY": true, "USD": true, "LKR": true}
		if !validCurrencies[*req.Currency] {
			vc.writeError(w, http.StatusBadRequest, "Invalid currency")
			return
		}
	}

	err = vc.vehicleService.UpdateVehicleDetails(r.Context(), id, &req)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle details updated successfully"})
}

func (vc *VehicleController) uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := r.ParseMultipartForm(32 << 20) // Increased size limit for multiple files

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid make ID")
		return
	}

	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get all files from the "images" field
	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "No images provided", http.StatusBadRequest)
		return
	}

	// Create upload directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return
	}

	var uploadedImages []entity.VehicleImage
	var errors []string

	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Unable to open file %s: %v", fileHeader.Filename, err))
			continue
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if !util.IsValidImageType(contentType) {
			file.Close()
			errors = append(errors, fmt.Sprintf("Invalid file type for %s. Only JPEG, PNG, GIF allowed", fileHeader.Filename))
			continue
		}

		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
		filePath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(filePath)
		if err != nil {
			file.Close()
			errors = append(errors, fmt.Sprintf("Unable to create file for %s: %v", fileHeader.Filename, err))
			continue
		}

		_, err = io.Copy(dst, file)
		file.Close()
		dst.Close()

		if err != nil {
			os.Remove(filePath)
			errors = append(errors, fmt.Sprintf("Unable to save file %s: %v", fileHeader.Filename, err))
			continue
		}

		var vehicleImage entity.VehicleImage
		vehicleImage.VehicleID = id
		vehicleImage.Filename = filename
		vehicleImage.FilePath = filePath
		vehicleImage.FileSize = fileHeader.Size
		vehicleImage.MimeType = contentType
		vehicleImage.DisplayOrder = i + 1
		vehicleImage.IsPrimary = i == 0
		vehicleImage.UploadDate = time.Now()
		if err != nil {
			// Clean up the file if database insert fails
			os.Remove(filePath)
			errors = append(errors, fmt.Sprintf("Database error for %s: %v", fileHeader.Filename, err))
			continue
		}

		uploadedImages = append(uploadedImages, vehicleImage)
	}

	images, err := vc.vehicleService.InsertVehicleImage(r.Context(), uploadedImages)
	// Prepare response
	response := map[string]interface{}{
		"uploaded_images": images,
		"total_uploaded":  len(images),
		"total_files":     len(files),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["partial_success"] = true
	}

	// Return appropriate status code
	if len(uploadedImages) == 0 {
		// All uploads failed
		vc.writeJSON(w, http.StatusBadRequest, response)
	} else if len(errors) > 0 {
		// Partial success
		vc.writeJSON(w, http.StatusMultiStatus, response)
	} else {
		// Complete success
		vc.writeJSON(w, http.StatusCreated, response)
	}
}

func (vc *VehicleController) serveImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Security: validate filename to prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("uploads", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}
