package controllers

import (
	//"car_service/dto/request"
	"car_service/filters"
	"car_service/services"
	"net/http"

	//"car_service/entity"
	//"car_service/util"
	//"database/sql"
	"encoding/json"
	//"fmt"
	//"io"
	//"os"
	//"path/filepath"
	//"strconv"
	//"strings"
	//"time"

	//"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AnalyticsController struct {
	analytics *services.AnalyticsService
	router    *mux.Router
}

func NewAnalyticController(analyticService *services.AnalyticsService, router *mux.Router) *AnalyticsController {
	return &AnalyticsController{
		analytics: analyticService,
		router:    router,
	}
}

func (ac *AnalyticsController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (ac *AnalyticsController) writeError(w http.ResponseWriter, status int, message string) {
	ac.writeJSON(w, status, map[string]string{"error": message})
}

func (ac *AnalyticsController) SetupRoutes() {
	api := ac.router.PathPrefix("/car-service/api/v1").Subrouter()

	vehicles := api.PathPrefix("/analytics").Subrouter()
	vehicles.HandleFunc("/shipping-status", ac.getShippingStatusCount).Methods("GET")
	vehicles.HandleFunc("/sales-status", ac.getSalesStatusCount).Methods("GET")
	vehicles.HandleFunc("/vehicle-brand-status", ac.getVehicleBrandCount).Methods("GET")
	vehicles.HandleFunc("/financial-summary", ac.getFiancialDetails).Methods("GET")
}

func (ac *AnalyticsController) getShippingStatusCount(w http.ResponseWriter, r *http.Request) {

	vehicleFilter := filters.NewVehicleShippingFilter()
	vehicleFilter.GetValuesFromRequest(r)

	shipping_status, err := ac.analytics.GetShippingStatusCount(r.Context(), vehicleFilter)
	if err != nil {
		ac.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ac.writeJSON(w, http.StatusOK, shipping_status)
}

func (ac *AnalyticsController) getSalesStatusCount(w http.ResponseWriter, r *http.Request) {

	vehicleSalesFilter := filters.NewVehicleSalesFilter()
	vehicleSalesFilter.GetValuesFromRequest(r)

	shipping_status, err := ac.analytics.GetSalesStatusCount(r.Context(), vehicleSalesFilter)
	if err != nil {
		ac.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ac.writeJSON(w, http.StatusOK, shipping_status)
}

func (ac *AnalyticsController) getVehicleBrandCount(w http.ResponseWriter, r *http.Request) {

	vehicleFilter := filters.NewVehicleFilters()
	vehicleFilter.GetValuesFromRequest(r)

	vehicle_brand_status, err := ac.analytics.GetVehicleBrandStatusCount(r.Context(), vehicleFilter)
	if err != nil {
		ac.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ac.writeJSON(w, http.StatusOK, vehicle_brand_status)
}

func (ac *AnalyticsController) getFiancialDetails(w http.ResponseWriter, r *http.Request) {

	vehicleFiancialFilter := filters.NewVehicleFinancialFilter()
	vehicleFiancialFilter.GetValuesFromRequest(r)

	financial_summary, err := ac.analytics.GetFinancialSummary(r.Context(), vehicleFiancialFilter)
	if err != nil {
		ac.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ac.writeJSON(w, http.StatusOK, financial_summary)
}
