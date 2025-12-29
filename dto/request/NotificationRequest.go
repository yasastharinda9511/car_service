package request

type NotificationRequest struct {
	NotificationType string                 `json:"notification_type"`
	Source           string                 `json:"source"`
	Payload          map[string]interface{} `json:"payload"`
	Priority         string                 `json:"priority"`
	ReferenceID      string                 `json:"reference_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
