package server

import (
	"car_service/config"
	"car_service/server/controllers"

	"car_service/services"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type APIServer struct {
	router *mux.Router
	db     *sql.DB
}

func NewAPIServer(db *sql.DB, cfg *config.Config) *APIServer {
	server := &APIServer{
		router: mux.NewRouter(),
		db:     db,
	}

	vehicleService := services.NewVehicleService(db)
	analyticService := services.NewAnalyticsService(db)

	// Initialize S3 service if enabled
	var s3Service *services.S3Service
	if cfg.UseS3Storage {
		var err error
		s3Service, err = services.NewS3Service(cfg.S3BucketName, cfg.S3Region)
		if err != nil {
			log.Printf("Warning: Failed to initialize S3 service: %v. Falling back to local storage.", err)
			cfg.UseS3Storage = false
		} else {
			log.Printf("S3 storage enabled: bucket=%s, region=%s", cfg.S3BucketName, cfg.S3Region)
		}
	} else {
		log.Println("Using local file storage for images")
	}

	vehicleController := controllers.NewVehicleController(vehicleService, s3Service, server.router, cfg.IntrospectURL)
	analyticController := controllers.NewAnalyticController(analyticService, server.router)
	vehicleMakeController := controllers.NewVehicleMakeController(server.router, cfg.IntrospectURL)

	vehicleController.SetupRoutes()
	analyticController.SetupRoutes()
	vehicleMakeController.SetupRoutes(db)

	server.setupRoutes()
	return server
}

// CORS Middleware
func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *APIServer) setupRoutes() {
	// Health check endpoints:
	// - /health: Comprehensive health check with database connectivity and schema verification
	// - /health/live: Liveness probe - returns 200 if the application is running
	// - /health/ready: Readiness probe - returns 200 if the application is ready to serve traffic
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
	s.router.HandleFunc("/health/live", s.livenessCheck).Methods("GET")
	s.router.HandleFunc("/health/ready", s.readinessCheck).Methods("GET")
}

// Helper functions for JSON responses
func (s *APIServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *APIServer) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// Health check handler
func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Initialize response
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "car-service-api",
		"version":   "1.0.0",
		"checks":    make(map[string]interface{}),
	}

	overallHealthy := true

	// Check database connectivity
	dbHealth := map[string]interface{}{
		"status": "healthy",
	}

	// Ping database with timeout
	if err := s.db.PingContext(ctx); err != nil {
		dbHealth["status"] = "unhealthy"
		dbHealth["error"] = err.Error()
		overallHealthy = false
	} else {
		// Get database stats
		stats := s.db.Stats()
		dbHealth["open_connections"] = stats.OpenConnections
		dbHealth["in_use"] = stats.InUse
		dbHealth["idle"] = stats.Idle
		dbHealth["max_open_connections"] = stats.MaxOpenConnections

		// Test a simple query to ensure schema access
		var schemaExists bool
		err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'cars')").Scan(&schemaExists)
		if err != nil {
			dbHealth["schema_check"] = "error"
			dbHealth["schema_error"] = err.Error()
			overallHealthy = false
		} else if !schemaExists {
			dbHealth["schema_check"] = "warning"
			dbHealth["schema_warning"] = "cars schema not found"
		} else {
			dbHealth["schema_check"] = "ok"
		}
	}

	response["checks"].(map[string]interface{})["database"] = dbHealth

	// Set overall status
	if !overallHealthy {
		response["status"] = "unhealthy"
		s.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Liveness check handler - checks if the application is alive
// This is used by Kubernetes to know if it should restart the pod
func (s *APIServer) livenessCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "car-service-api",
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Readiness check handler - checks if the application is ready to serve traffic
// This is used by Kubernetes to know if it should send traffic to the pod
func (s *APIServer) readinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "car-service-api",
	}

	// Check database connectivity for readiness
	if err := s.db.PingContext(ctx); err != nil {
		response["status"] = "not_ready"
		response["reason"] = "database_unavailable"
		response["error"] = err.Error()
		s.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	// Verify cars schema exists
	var schemaExists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'cars')").Scan(&schemaExists)
	if err != nil || !schemaExists {
		response["status"] = "not_ready"
		response["reason"] = "schema_not_found"
		if err != nil {
			response["error"] = err.Error()
		}
		s.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) Start(port string, allowedOrigins []string) error {
	log.Printf("Starting server on port %s", port)
	cors := config.NewCorsConfig(allowedOrigins)
	return http.ListenAndServe(":"+port, cors.WithCORS(s.router))
}

//	func (s *APIServer) createVehicleMake(w http.ResponseWriter, r *http.Request) {
//		var req request.CreateVehicleMake
//
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid JSON")
//			return
//		}
//
//		// Validation
//		if *req.MakeName == "" {
//			s.writeError(w, http.StatusBadRequest, "Make name is required")
//			return
//		}
//
//		// Set defaults
//		if req.CountryOrigin == nil || *req.CountryOrigin == "" {
//			defaultVal := "Japan"
//			req.CountryOrigin = &defaultVal
//		}
//		if req.IsActive == nil {
//			defaultActive := true
//			req.IsActive = &defaultActive
//		}
//
//		make, err := s.vehicleService.CreateVehicleMake(req)
//		if err != nil {
//			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
//				s.writeError(w, http.StatusConflict, "Make name already exists")
//				return
//			}
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusCreated, map[string]interface{}{
//			"data":    make,
//			"message": "Vehicle make created successfully",
//		})
//	}
//func (s *APIServer) getVehicleMakes(w http.ResponseWriter, r *http.Request) {
//	activeOnly := r.URL.Query().Get("active_only") == "true"
//
//	makes, err := s.vehicleService.GetAllVehicleMakes(activeOnly)
//	if err != nil {
//		s.writeError(w, http.StatusInternalServerError, err.Error())
//		return
//	}
//
//	s.writeJSON(w, http.StatusOK, map[string]interface{}{
//		"data": makes,
//		"meta": map[string]interface{}{
//			"total":       len(makes),
//			"active_only": activeOnly,
//		},
//	})
//}

//	func (s *APIServer) updateVehicleMake(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.Atoi(vars["id"])
//		if err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid make ID")
//			return
//		}
//
//		var req request.CreateVehicleMake
//
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid JSON")
//			return
//		}
//
//		err = s.vehicleService.UpdateVehicleMake(id, req.MakeName, req.CountryOrigin, req.IsActive)
//		if err != nil {
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle make updated successfully"})
//	}
//
//	func (s *APIServer) createVehicleModel(w http.ResponseWriter, r *http.Request) {
//		var req request.CreateVehicleModel
//
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid JSON")
//			return
//		}
//
//		// Validation
//		if req.MakeID <= 0 {
//			s.writeError(w, http.StatusBadRequest, "Valid make ID is required")
//			return
//		}
//		if *req.ModelName == "" {
//			s.writeError(w, http.StatusBadRequest, "Model name is required")
//			return
//		}
//
//		// Set defaults
//		if req.IsActive == nil {
//			defaultActive := true
//			req.IsActive = &defaultActive
//		}
//
//		model, err := s.vehicleService.CreateVehicleModel(req)
//		if err != nil {
//			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
//				s.writeError(w, http.StatusConflict, "Model name already exists for this make")
//				return
//			}
//			if strings.Contains(err.Error(), "foreign key") {
//				s.writeError(w, http.StatusBadRequest, "Invalid make ID")
//				return
//			}
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusCreated, map[string]interface{}{
//			"data":    model,
//			"message": "Vehicle model created successfully",
//		})
//	}
//
//	func (s *APIServer) getVehicleModels(w http.ResponseWriter, r *http.Request) {
//		makeIDStr := r.URL.Query().Get("make_id")
//		activeOnly := r.URL.Query().Get("active_only") == "true"
//
//		var makeID *int
//		if makeIDStr != "" {
//			if id, err := strconv.Atoi(makeIDStr); err == nil {
//				makeID = &id
//			} else {
//				s.writeError(w, http.StatusBadRequest, "Invalid make_id parameter")
//				return
//			}
//		}
//
//		models, err := s.vehicleService.GetVehicleModels(makeID, activeOnly)
//		if err != nil {
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusOK, map[string]interface{}{
//			"data": models,
//			"meta": map[string]interface{}{
//				"total":       len(models),
//				"make_id":     makeID,
//				"active_only": activeOnly,
//			},
//		})
//	}
//
//	func (s *APIServer) getVehicleModel(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.Atoi(vars["id"])
//		if err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid model ID")
//			return
//		}
//
//		model, err := s.vehicleService.GetVehicleModelByID(id)
//		if err != nil {
//			if err == sql.ErrNoRows {
//				s.writeError(w, http.StatusNotFound, "Vehicle model not found")
//				return
//			}
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": model})
//	}
//
//	func (s *APIServer) updateVehicleModel(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.Atoi(vars["id"])
//		if err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid model ID")
//			return
//		}
//
//		var req request.CreateVehicleModel
//
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid JSON")
//			return
//		}
//
//		err = s.vehicleService.UpdateVehicleModel(id, req.ModelName, req.BodyType,
//			req.FuelType, req.TransmissionType, req.EngineSizeCC, req.IsActive)
//		if err != nil {
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusOK, map[string]string{"message": "Vehicle model updated successfully"})
//	}
//
// // Order handlers
//
//	func (s *APIServer) getOrders(w http.ResponseWriter, r *http.Request) {
//		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
//		if page < 1 {
//			page = 1
//		}
//		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
//		if limit < 1 {
//			limit = 10
//		}
//		offset := (page - 1) * limit
//
//		orders, err := s.orderService.GetAllOrders(limit, offset)
//		if err != nil {
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		response := map[string]interface{}{
//			"data": orders,
//			"meta": map[string]interface{}{
//				"page":  page,
//				"limit": limit,
//				"total": len(orders),
//			},
//		}
//		s.writeJSON(w, http.StatusOK, response)
//	}
//
//	func (s *APIServer) getOrder(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.ParseInt(vars["id"], 10, 64)
//		if err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid order ID")
//			return
//		}
//
//		order, err := s.orderService.GetOrderByID(id)
//		if err != nil {
//			if err == sql.ErrNoRows {
//				s.writeError(w, http.StatusNotFound, "Order not found")
//				return
//			}
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": order})
//	}
//
//	func (s *APIServer) createOrder(w http.ResponseWriter, r *http.Request) {
//		var req request.CreateOrderRequest
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid JSON")
//			return
//		}
//
//		// Validation
//		if req.CustomerName == "" || req.ContactNumber == "" || req.PreferredMake == "" || req.PreferredModel == "" {
//			s.writeError(w, http.StatusBadRequest, "Missing required fields")
//			return
//		}
//
//		order, err := s.orderService.CreateOrder(req)
//		if err != nil {
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		response := map[string]interface{}{
//			"data":    order,
//			"message": "Order created successfully",
//		}
//		s.writeJSON(w, http.StatusCreated, response)
//	}
//
//	func (s *APIServer) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.ParseInt(vars["id"], 10, 64)
//		if err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid order ID")
//			return
//		}
//
//		var req struct {
//			Status string `json:"status"`
//		}
//		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//			s.writeError(w, http.StatusBadRequest, "Invalid JSON")
//			return
//		}
//
//		if req.Status == "" {
//			s.writeError(w, http.StatusBadRequest, "Status is required")
//			return
//		}
//
//		err = s.orderService.UpdateOrderStatus(id, req.Status)
//		if err != nil {
//			s.writeError(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		s.writeJSON(w, http.StatusOK, map[string]string{"message": "Order status updated successfully"})
//	}
//
// // Analytics handlers
//
//	func (s *APIServer) getDashboardStats(w http.ResponseWriter, r *http.Request) {
//		// This would typically aggregate data from multiple tables
//		stats := map[string]interface{}{
//			"total_vehicles":    150,
//			"vehicles_in_stock": 45,
//			"vehicles_sold":     105,
//			"pending_orders":    8,
//			"total_revenue":     450000000, // LKR
//			"total_profit":      75000000,  // LKR
//			"profit_margin":     16.67,     // %
//		}
//		s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": stats})
//	}
//
//	func (s *APIServer) getSalesSummary(w http.ResponseWriter, r *http.Request) {
//		// Monthly sales summary - would typically come from database
//		summary := []map[string]interface{}{
//			{
//				"month":         "2024-09",
//				"vehicles_sold": 12,
//				"total_revenue": 45000000,
//				"total_profit":  7500000,
//				"profit_margin": 16.67,
//			},
//			{
//				"month":         "2024-08",
//				"vehicles_sold": 15,
//				"total_revenue": 52000000,
//				"total_profit":  8200000,
//				"profit_margin": 15.77,
//			},
//		}
//		s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": summary})
//	}
//
//	func (s *APIServer) getInventoryStatus(w http.ResponseWriter, r *http.Request) {
//		// Inventory breakdown by status
//		inventory := []map[string]interface{}{
//			{
//				"shipping_status": "DELIVERED",
//				"sale_status":     "SOLD",
//				"count":           105,
//				"investment":      315000000,
//			},
//			{
//				"shipping_status": "CLEARED",
//				"sale_status":     "AVAILABLE",
//				"count":           25,
//				"investment":      75000000,
//			},
//			{
//				"shipping_status": "ARRIVED",
//				"sale_status":     "AVAILABLE",
//				"count":           15,
//				"investment":      45000000,
//			},
//			{
//				"shipping_status": "SHIPPED",
//				"sale_status":     "AVAILABLE",
//				"count":           5,
//				"investment":      15000000,
//			},
//		}
//		s.writeJSON(w, http.StatusOK, map[string]interface{}{"data": inventory})
//	}
