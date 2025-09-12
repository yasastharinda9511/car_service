package response

type DetailedFinancialSummary struct {
	TotalCharges       float64 `json:"total_charges" db:"total_charges"`
	TotalTT            float64 `json:"total_tt" db:"total_tt"`
	TotalDuty          float64 `json:"total_duty" db:"total_duty"`
	TotalClearing      float64 `json:"total_clearing" db:"total_clearing"`
	TotalOtherExpenses float64 `json:"total_other_expenses" db:"total_other_expenses"`
	TotalInvestment    float64 `json:"total_investment" db:"total_investment"`
}
