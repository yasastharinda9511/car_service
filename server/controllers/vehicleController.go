package controllers

import (
	"car_service/dto/request"
	"car_service/filters"
	"car_service/internal/constants"
	"car_service/middleware"
	"car_service/services"
	"net/http"
	"path"

	"car_service/entity"
	"car_service/util"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type VehicleController struct {
	vehicleService *services.VehicleService
	s3Service      *services.S3Service
	router         *mux.Router
	introspectURL  string
}

func NewVehicleController(vehicleService *services.VehicleService, s3Service *services.S3Service, router *mux.Router, introspectURL string) *VehicleController {
	return &VehicleController{
		vehicleService: vehicleService,
		s3Service:      s3Service,
		router:         router,
		introspectURL:  introspectURL,
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
	authMiddleware := middleware.NewAuthMiddleware(vc.introspectURL)

	// Vehicle routes
	vehicles := api.PathPrefix("/vehicles").Subrouter()

	vehicles.Handle("", authMiddleware.Authorize(http.HandlerFunc(vc.getVehicles), constants.VEHICLE_ACCESS)).Methods("GET")
	vehicles.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.getVehicle), constants.VEHICLE_ACCESS)).Methods("GET")
	vehicles.Handle("", authMiddleware.Authorize(http.HandlerFunc(vc.createVehicle), constants.VEHICLE_CREATE)).Methods("POST")
	vehicles.Handle("/download-image/{id}/{filename}", authMiddleware.Authorize(http.HandlerFunc(vc.serveImageHandler), constants.VEHICLE_ACCESS)).Methods("GET")
	vehicles.Handle("/upload-image/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.uploadImageHandler), constants.VEHICLE_CREATE)).Methods("POST")
	vehicles.Handle("/{id}/images/{imageId}/set-primary", authMiddleware.Authorize(http.HandlerFunc(vc.setPrimaryImageHandler), constants.VEHICLE_EDIT)).Methods("PUT")

	vehicles.Handle("/{id}/shipping", authMiddleware.Authorize(http.HandlerFunc(vc.updateShipping), constants.SHIPPING_EDIT)).Methods("PUT")
	vehicles.Handle("/{id}/purchase", authMiddleware.Authorize(http.HandlerFunc(vc.updatePurchase), constants.PURCHASE_EDIT)).Methods("PUT")
	vehicles.Handle("/{id}/financials", authMiddleware.Authorize(http.HandlerFunc(vc.updateFinancials), constants.FINANCIAL_EDIT)).Methods("PUT")
	vehicles.Handle("/{id}/sales", authMiddleware.Authorize(http.HandlerFunc(vc.updateSales), constants.SALES_EDIT)).Methods("PUT")
	vehicles.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.updateVehicle), constants.VEHICLE_EDIT)).Methods("PUT")
	vehicles.Handle("/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.deleteVehicle), constants.VEHICLE_DELETE)).Methods("DELETE")

	// Dropdown data route
	vehicles.HandleFunc("/dropdown/options", vc.getDropdownOptions).Methods("GET")

	// Customer management routes
	vehicles.Handle("/{id}/customer", authMiddleware.Authorize(http.HandlerFunc(vc.assignCustomer), constants.SALES_EDIT)).Methods("PUT")
	vehicles.Handle("/{id}/customer", authMiddleware.Authorize(http.HandlerFunc(vc.removeCustomer), constants.SALES_EDIT)).Methods("DELETE")
	vehicles.Handle("/customer/{customer_id}", authMiddleware.Authorize(http.HandlerFunc(vc.getVehiclesByCustomer), constants.VEHICLE_ACCESS)).Methods("GET")

	// Shipping history routes
	vehicles.Handle("/shipping/history/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.getShippingHistory), constants.SHIIPING_ACCESS)).Methods("GET")
	vehicles.Handle("/shipping/history/recent", authMiddleware.Authorize(http.HandlerFunc(vc.getRecentShippingHistory), constants.SHIIPING_ACCESS)).Methods("GET")

	// Purchase history routes
	vehicles.Handle("/purchase/history/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.getPurchaseHistory), constants.PURCHASE_ACCESS)).Methods("GET")
	vehicles.Handle("/purchase/history/recent", authMiddleware.Authorize(http.HandlerFunc(vc.getRecentPurchaseHistory), constants.PURCHASE_ACCESS)).Methods("GET")
	vehicles.Handle("/purchase/history/status/{status}", authMiddleware.Authorize(http.HandlerFunc(vc.getPurchaseHistoryByStatus), constants.PURCHASE_ACCESS)).Methods("GET")
	vehicles.Handle("/purchase/history/supplier/{supplier_id}", authMiddleware.Authorize(http.HandlerFunc(vc.getPurchaseHistoryBySupplier), constants.PURCHASE_ACCESS)).Methods("GET")

	// Document management routes
	vehicles.Handle("/upload-document/{id}", authMiddleware.Authorize(http.HandlerFunc(vc.uploadDocumentHandler), constants.VEHICLE_CREATE)).Methods("POST")
	vehicles.Handle("/download-document/{document_id}", authMiddleware.Authorize(http.HandlerFunc(vc.serveDocumentHandler), constants.VEHICLE_ACCESS)).Methods("GET")

	// Featured vehicles routes
	vehicles.Handle("/featured", authMiddleware.Authorize(http.HandlerFunc(vc.getFeaturedVehicles), constants.VEHICLE_ACCESS)).Methods("GET")
	vehicles.Handle("/{id}/featured", authMiddleware.Authorize(http.HandlerFunc(vc.setVehicleFeatured), constants.VEHICLE_EDIT)).Methods("PUT")

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

	// Extract authorization header
	authHeader := r.Header.Get("Authorization")

	vehicle, err := vc.vehicleService.CreateVehicle(r.Context(), req, authHeader)
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

	// Extract authorization header from request
	authHeader := r.Header.Get("Authorization")

	err = vc.vehicleService.UpdateShippingStatus(r.Context(), id, req, authHeader)
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

	// Extract authorization header from request
	authHeader := r.Header.Get("Authorization")

	err = vc.vehicleService.UpdatePurchaseDetails(r.Context(), id, &req, authHeader)
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

	// Note: TotalCostLKR is now auto-calculated including LC cost

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

	// Validation: If status is SOLD, require customer and revenue
	if req.SaleStatus == "SOLD" {
		if req.CustomerID == nil || *req.CustomerID <= 0 {
			vc.writeError(w, http.StatusBadRequest, "Customer ID is required when status is SOLD")
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
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		http.Error(w, "No images provided", http.StatusBadRequest)
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

		var vehicleImage entity.VehicleImage
		vehicleImage.VehicleID = int64(id)
		vehicleImage.OriginalName = fileHeader.Filename
		vehicleImage.FileSize = fileHeader.Size
		vehicleImage.MimeType = contentType
		vehicleImage.DisplayOrder = i + 1
		vehicleImage.IsPrimary = i == 0
		vehicleImage.UploadDate = time.Now()

		pathPrefix := fmt.Sprintf("vehicles/%d/images", id)

		if vc.s3Service != nil {
			result, err := vc.s3Service.UploadFile(r.Context(), file, fileHeader, pathPrefix)
			file.Close()

			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to upload %s to S3: %v", fileHeader.Filename, err))
				continue
			}

			vehicleImage.Filename = result.Filename
			vehicleImage.FilePath = result.Key // Store S3 key in file_path
		}

		uploadedImages = append(uploadedImages, vehicleImage)
	}

	images, err := vc.vehicleService.InsertVehicleImage(r.Context(), uploadedImages)

	// Prepare response
	response := map[string]interface{}{
		"uploaded_images": images,
		"total_uploaded":  len(images),
		"total_files":     len(files),
		"storage_type":    "s3",
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
	id := vars["id"]

	// Security: validate filename to prevent directory traversal
	if strings.Contains(filename, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	if vc.s3Service != nil {
		// For S3, generate a presigned URL and redirect
		// The S3 key would be stored in the database as file_path
		// For this endpoint, we construct the key from the filename
		pathPrefix := fmt.Sprintf("vehicles/%s/images", id)
		key := path.Join(pathPrefix, filename)

		//// Check if file exists in S3
		exists, err := vc.s3Service.CheckIfFileExists(r.Context(), key)
		if err != nil || !exists {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		// Generate presigned URL (valid for 15 minutes)
		presignedURL, err := vc.s3Service.GetPresignedURL(r.Context(), key, 15)
		if err != nil {
			http.Error(w, "Failed to generate image URL", http.StatusInternalServerError)
			return
		}

		// Redirect to the presigned URL
		vc.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": presignedURL,
		})
	}
}

func (vc *VehicleController) setPrimaryImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	imageID, err := strconv.Atoi(vars["imageId"])
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid image ID")
		return
	}

	err = vc.vehicleService.SetPrimaryImage(r.Context(), imageID, vehicleID)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Primary image set successfully",
	})
}

func (vc *VehicleController) uploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := r.ParseMultipartForm(50 << 20) // 50MB limit for documents

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["documents"]
	if len(files) == 0 {
		http.Error(w, "No documents provided", http.StatusBadRequest)
		return
	}

	// Get document_type from form data
	documentTypes := r.MultipartForm.Value["document_type"]
	documentNames := r.MultipartForm.Value["document_name"]

	var uploadedDocuments []entity.VehicleDocument
	var errors []string

	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Unable to open file %s: %v", fileHeader.Filename, err))
			continue
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if !util.IsValidDocumentType(contentType) {
			file.Close()
			errors = append(errors, fmt.Sprintf("Invalid file type for %s. Only PDF, images, and Office documents allowed", fileHeader.Filename))
			continue
		}

		// Get document type for this file (default to OTHER if not provided)
		var docType entity.DocumentType = entity.DocumentTypeOther
		if i < len(documentTypes) && documentTypes[i] != "" {
			docType = entity.DocumentType(documentTypes[i])
		}

		// Get document name for this file (default to filename if not provided)
		docName := fileHeader.Filename
		if i < len(documentNames) && documentNames[i] != "" {
			docName = documentNames[i]
		}

		var vehicleDocument entity.VehicleDocument
		vehicleDocument.VehicleID = int64(id)
		vehicleDocument.DocumentType = docType
		vehicleDocument.DocumentName = docName
		vehicleDocument.FileSizeBytes = fileHeader.Size
		vehicleDocument.MimeType = contentType
		vehicleDocument.UploadDate = time.Now()

		pathPrefix := fmt.Sprintf("vehicles/%d/documents", id)

		if vc.s3Service != nil {
			result, err := vc.s3Service.UploadFile(r.Context(), file, fileHeader, pathPrefix)
			file.Close()

			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to upload %s to S3: %v", fileHeader.Filename, err))
				fmt.Println("failed to upload file" + fileHeader.Filename + ": " + err.Error())
				continue
			}

			vehicleDocument.FilePath = result.Key // Store S3 key in file_path
		}

		uploadedDocuments = append(uploadedDocuments, vehicleDocument)
	}

	documents, err := vc.vehicleService.InsertVehicleDocument(r.Context(), uploadedDocuments)

	// Prepare response
	response := map[string]interface{}{
		"uploaded_documents": documents,
		"total_uploaded":     len(documents),
		"total_files":        len(files),
		"storage_type":       "s3",
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["partial_success"] = true
	}

	// Return appropriate status code
	if len(uploadedDocuments) == 0 {
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

func (vc *VehicleController) serveDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	documentID, err := strconv.ParseInt(vars["document_id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Get document metadata from database
	document, err := vc.vehicleService.GetVehicleDocumentByID(r.Context(), documentID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve document", http.StatusInternalServerError)
		return
	}

	if vc.s3Service != nil {
		// Check if file exists in S3
		exists, err := vc.s3Service.CheckIfFileExists(r.Context(), document.FilePath)
		if err != nil || !exists {
			http.Error(w, "Document not found in storage", http.StatusNotFound)
			return
		}

		// Generate presigned URL (valid for 15 minutes)
		presignedURL, err := vc.s3Service.GetPresignedURL(r.Context(), document.FilePath, 15)
		if err != nil {
			http.Error(w, "Failed to generate document URL", http.StatusInternalServerError)
			return
		}

		// Return the presigned URL along with document metadata
		vc.writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": presignedURL,
			"metadata": map[string]interface{}{
				"document_name": document.DocumentName,
				"document_type": document.DocumentType,
				"mime_type":     document.MimeType,
				"file_size":     document.FileSizeBytes,
			},
		})
	}
}

func (vc *VehicleController) deleteVehicle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	// Extract authorization header from request
	authHeader := r.Header.Get("Authorization")

	err = vc.vehicleService.DeleteVehicle(r.Context(), vehicleID, authHeader)
	if err != nil {
		if err == sql.ErrNoRows {
			vc.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		vc.writeError(w, http.StatusInternalServerError, "Failed to delete vehicle")
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Vehicle deleted successfully",
	})
}

// getDropdownOptions returns all distinct values for dropdown filters in a single call
func (vc *VehicleController) getDropdownOptions(w http.ResponseWriter, r *http.Request) {
	options, err := vc.vehicleService.GetDropdownOptions(r.Context())
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": options,
	})
}

// assignCustomer assigns a customer to a vehicle sale
func (vc *VehicleController) assignCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var req struct {
		CustomerID int64 `json:"customer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.CustomerID <= 0 {
		vc.writeError(w, http.StatusBadRequest, "Valid customer ID is required")
		return
	}

	err = vc.vehicleService.AssignCustomerToVehicle(r.Context(), vehicleID, req.CustomerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			vc.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "foreign key") {
			vc.writeError(w, http.StatusBadRequest, "Invalid customer ID")
			return
		}
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Customer assigned to vehicle successfully",
	})
}

// removeCustomer removes the customer assignment from a vehicle sale
func (vc *VehicleController) removeCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	err = vc.vehicleService.RemoveCustomerFromVehicle(r.Context(), vehicleID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			vc.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Customer removed from vehicle successfully",
	})
}

// getVehiclesByCustomer retrieves all vehicles associated with a specific customer
func (vc *VehicleController) getVehiclesByCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := strconv.ParseInt(vars["customer_id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	vehicles, err := vc.vehicleService.GetVehiclesByCustomer(r.Context(), customerID)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": vehicles,
		"meta": map[string]interface{}{
			"customer_id": customerID,
			"total":       len(vehicles),
		},
	})
}

// getShippingHistory retrieves the shipping status change history for a specific vehicle
func (vc *VehicleController) getShippingHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	history, err := vc.vehicleService.GetShippingHistory(r.Context(), vehicleID)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
		"meta": map[string]interface{}{
			"vehicle_id": vehicleID,
			"total":      len(history),
		},
	})
}

// getRecentShippingHistory retrieves recent shipping status changes across all vehicles
func (vc *VehicleController) getRecentShippingHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 200 {
		limit = 200 // Max limit
	}

	history, err := vc.vehicleService.GetRecentShippingHistory(r.Context(), limit)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
		"meta": map[string]interface{}{
			"limit": limit,
			"total": len(history),
		},
	})
}

// getPurchaseHistory retrieves the purchase change history for a specific vehicle
func (vc *VehicleController) getPurchaseHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	history, err := vc.vehicleService.GetPurchaseHistory(r.Context(), vehicleID)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
		"meta": map[string]interface{}{
			"vehicle_id": vehicleID,
			"total":      len(history),
		},
	})
}

// getRecentPurchaseHistory retrieves recent purchase changes across all vehicles
func (vc *VehicleController) getRecentPurchaseHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 200 {
		limit = 200 // Max limit
	}

	history, err := vc.vehicleService.GetRecentPurchaseHistory(r.Context(), limit)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
		"meta": map[string]interface{}{
			"limit": limit,
			"total": len(history),
		},
	})
}

// getPurchaseHistoryByStatus retrieves purchase history filtered by status
func (vc *VehicleController) getPurchaseHistoryByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status := vars["status"]

	// Validate status enum
	validStatuses := map[string]bool{
		"LC_PENDING": true, "LC_OPENED": true, "LC_RECEIVED": true, "CANCELLED": true,
	}
	if !validStatuses[status] {
		vc.writeError(w, http.StatusBadRequest, "Invalid purchase status")
		return
	}

	history, err := vc.vehicleService.GetPurchaseHistoryByStatus(r.Context(), status)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
		"meta": map[string]interface{}{
			"status": status,
			"total":  len(history),
		},
	})
}

// getPurchaseHistoryBySupplier retrieves purchase history for a specific supplier
func (vc *VehicleController) getPurchaseHistoryBySupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	supplierID, err := strconv.ParseInt(vars["supplier_id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid supplier ID")
		return
	}

	history, err := vc.vehicleService.GetPurchaseHistoryBySupplier(r.Context(), supplierID)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
		"meta": map[string]interface{}{
			"supplier_id": supplierID,
			"total":       len(history),
		},
	})
}

// setVehicleFeatured marks a vehicle as featured or unfeatured
func (vc *VehicleController) setVehicleFeatured(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid vehicle ID")
		return
	}

	var req struct {
		IsFeatured bool `json:"is_featured"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vc.writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Extract authorization header
	authHeader := r.Header.Get("Authorization")

	err = vc.vehicleService.SetVehicleFeatured(r.Context(), vehicleID, req.IsFeatured, authHeader)
	if err != nil {
		if err == sql.ErrNoRows {
			vc.writeError(w, http.StatusNotFound, "Vehicle not found")
			return
		}
		vc.writeError(w, http.StatusInternalServerError, "Failed to update featured status")
		return
	}

	statusMsg := "unfeatured"
	if req.IsFeatured {
		statusMsg = "featured"
	}

	vc.writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Vehicle %s successfully", statusMsg),
	})
}

// getFeaturedVehicles retrieves all featured vehicles
func (vc *VehicleController) getFeaturedVehicles(w http.ResponseWriter, r *http.Request) {
	// Get limit from query params (default to 10, max 50)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	vehicles, err := vc.vehicleService.GetFeaturedVehicles(r.Context(), limit)
	if err != nil {
		vc.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	vc.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": vehicles,
		"meta": map[string]interface{}{
			"total": len(vehicles),
			"limit": limit,
		},
	})
}
