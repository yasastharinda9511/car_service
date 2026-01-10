package filters

// VehicleFieldMapping maps user-friendly field names to database column names with aliases
var VehicleFieldMapping = map[string]string{
	// Vehicle table fields (alias: v)
	"make":                "v.make",
	"model":               "v.model",
	"trim_level":          "v.trim_level",
	"year":                "v.year_of_manufacture",
	"year_of_manufacture": "v.year_of_manufacture",
	"color":               "v.color",
	"mileage":             "v.mileage_km",
	"mileage_km":          "v.mileage_km",
	"chassis_id":          "v.chassis_id",
	"condition_status":    "v.condition_status",
	"auction_grade":       "v.auction_grade",
	"auction_price":       "v.auction_price",
	"price_quoted":        "v.price_quoted",
	"cif_value":           "v.cif_value",
	"currency":            "v.currency",
	"created_at":          "v.created_at",
	"updated_at":          "v.updated_at",
	"code":                "v.code",
	"id":                  "v.id",
	"is_featured":         "v.is_featured",
	"featured_at":         "v.featured_at",

	// Shipping table fields (alias: vs)
	"shipping_status":   "vs.shipping_status",
	"vessel_name":       "vs.vessel_name",
	"departure_harbour": "vs.departure_harbour",
	"shipment_date":     "vs.shipment_date",
	"arrival_date":      "vs.arrival_date",
	"clearing_date":     "vs.clearing_date",

	// Sales table fields (alias: vsl)
	"sale_status":      "vsl.sale_status",
	"sold_date":        "vsl.sold_date",
	"revenue":          "vsl.revenue",
	"profit":           "vsl.profit",
	"sold_to_name":     "vsl.sold_to_name",
	"customer_address": "vsl.customer_address",

	// Purchase table fields (alias: vp)
	"bought_from_name": "vp.bought_from_name",
	"purchase_date":    "vp.purchase_date",
	"lc_bank":          "vp.lc_bank",
	"lc_number":        "vp.lc_number",
	"lc_cost":          "vp.lc_cost_jpy",
	"purchase_status":  "vp.purchase_status",
	"supplier_id":      "vp.supplier_id",

	// Financial table fields (alias: vf)
	"total_cost":     "vf.total_cost_lkr",
	"charges":        "vf.charges_lkr",
	"duty":           "vf.duty_lkr",
	"clearing":       "vf.clearing_lkr",
	"other_expenses": "vf.other_expenses_lkr",
}

// GetMappedField returns the database field name with alias for a given user-friendly field name
// Returns the original field if no mapping exists
func GetMappedField(fieldName string) string {
	if mappedField, exists := VehicleFieldMapping[fieldName]; exists {
		return mappedField
	}
	return fieldName
}

// IsValidOrderByField checks if a field is valid for ordering
func IsValidOrderByField(fieldName string) bool {
	_, exists := VehicleFieldMapping[fieldName]
	return exists
}

// ValidOrderByFields returns a list of all valid field names that can be used for ordering
func ValidOrderByFields() []string {
	fields := make([]string, 0, len(VehicleFieldMapping))
	for field := range VehicleFieldMapping {
		fields = append(fields, field)
	}
	return fields
}
