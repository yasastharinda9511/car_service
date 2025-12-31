package server

import (
	"car_service/config"
	"car_service/logger"
	"car_service/server/controllers"
	"car_service/services"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type APIServer struct {
	router *mux.Router
	db     *sql.DB
}

func NewAPIServer(db *sql.DB, cfg *config.Config) *APIServer {
	logger.Info("Initializing API server")

	server := &APIServer{
		router: mux.NewRouter(),
		db:     db,
	}

	logger.Debug("Creating service instances")
	notificationService := services.NewNotificationService(cfg.NotificationServiceURL)
	vehicleService := services.NewVehicleService(db, notificationService)
	customerService := services.NewCustomerService(db, notificationService)
	supplierService := services.NewSupplierService(db, notificationService)
	analyticService := services.NewAnalyticsService(db)

	// Initialize S3 service if enabled
	var s3Service *services.S3Service
	if cfg.UseS3Storage {
		logger.WithFields(map[string]interface{}{
			"bucket": cfg.BucketName,
			"region": cfg.Region,
		}).Info("Initializing S3 storage service")

		var err error
		s3Service, err = services.NewS3Service(cfg.Region, cfg.AccessKey, cfg.SecretKey, cfg.BucketName)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"error":  err.Error(),
				"bucket": cfg.BucketName,
				"region": cfg.Region,
			}).Warn("Failed to initialize S3 service, falling back to local storage")
			cfg.UseS3Storage = false
		} else {
			logger.WithFields(map[string]interface{}{
				"bucket": cfg.BucketName,
				"region": cfg.Region,
			}).Info("S3 storage initialized successfully")
		}
	} else {
		logger.Info("Using local file storage for images")
	}

	logger.Debug("Initializing controllers")
	vehicleController := controllers.NewVehicleController(vehicleService, s3Service, server.router, cfg.IntrospectURL)
	analyticController := controllers.NewAnalyticController(analyticService, server.router)
	vehicleMakeController := controllers.NewVehicleMakeController(server.router, cfg.IntrospectURL, s3Service)
	vehicleModelController := controllers.NewVehicleModelController(server.router, cfg.IntrospectURL)
	customerController := controllers.NewCustomerController(server.router, cfg.IntrospectURL, customerService)
	supplierController := controllers.NewSupplierController(server.router, cfg.IntrospectURL, supplierService)

	logger.Debug("Setting up controller routes")
	vehicleController.SetupRoutes()
	analyticController.SetupRoutes()
	vehicleMakeController.SetupRoutes(db)
	vehicleModelController.SetupRoutes(db)
	customerController.SetupRoutes(db)
	supplierController.SetupRoutes(db)

	server.setupRoutes()
	logger.Info("API server initialization completed")
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
	logger.Debug("Health check requested")

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
		logger.WithField("error", err.Error()).Error("Health check: Database ping failed")
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
			logger.WithField("error", err.Error()).Error("Health check: Schema verification failed")
			dbHealth["schema_check"] = "error"
			dbHealth["schema_error"] = err.Error()
			overallHealthy = false
		} else if !schemaExists {
			logger.Warn("Health check: Cars schema not found")
			dbHealth["schema_check"] = "warning"
			dbHealth["schema_warning"] = "cars schema not found"
		} else {
			dbHealth["schema_check"] = "ok"
		}
	}

	response["checks"].(map[string]interface{})["database"] = dbHealth

	// Set overall status
	if !overallHealthy {
		logger.Warn("Health check failed")
		response["status"] = "unhealthy"
		s.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	logger.Debug("Health check passed")
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
	logger.Debug("Readiness check requested")

	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "car-service-api",
	}

	// Check database connectivity for readiness
	if err := s.db.PingContext(ctx); err != nil {
		logger.WithField("error", err.Error()).Warn("Readiness check: Database unavailable")
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
		logger.WithField("error", err).Warn("Readiness check: Schema not found or inaccessible")
		response["status"] = "not_ready"
		response["reason"] = "schema_not_found"
		if err != nil {
			response["error"] = err.Error()
		}
		s.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	logger.Debug("Readiness check passed")
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) Start(port string, allowedOrigins []string) error {
	logger.WithFields(map[string]interface{}{
		"port":            port,
		"allowed_origins": allowedOrigins,
	}).Info("Starting HTTP server")

	cors := config.NewCorsConfig(allowedOrigins)
	err := http.ListenAndServe(":"+port, cors.WithCORS(s.router))

	if err != nil {
		logger.WithFields(map[string]interface{}{
			"port":  port,
			"error": err.Error(),
		}).Error("Server failed to start")
	}

	return err
}
