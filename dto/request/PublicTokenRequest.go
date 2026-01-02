package request

type PublicTokenRequest struct {
	ExpireInDays   int      `json:"expire_in_days"`
	IncludeDetails []string `json:"include_details"`
}
