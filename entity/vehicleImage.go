package entity

import "time"

type VehicleImage struct {
	ID           int       `json:"id"`
	VehicleID    int64     `json:"vehicle_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	IsPrimary    bool      `json:"is_primary"`
	UploadDate   time.Time `json:"upload_date"`
	DisplayOrder int       `json:"display_order"`
}
