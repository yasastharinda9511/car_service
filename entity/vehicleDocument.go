package entity

import "time"

type DocumentType string

const (
	DocumentTypeInvoice      DocumentType = "INVOICE"
	DocumentTypeShipping     DocumentType = "SHIPPING"
	DocumentTypeCustoms      DocumentType = "CUSTOMS"
	DocumentTypeInspection   DocumentType = "INSPECTION"
	DocumentTypeRegistration DocumentType = "REGISTRATION"
	DocumentTypeOther        DocumentType = "OTHER"
)

type VehicleDocument struct {
	ID            int64        `json:"id"`
	VehicleID     int64        `json:"vehicle_id"`
	DocumentType  DocumentType `json:"document_type"`
	DocumentName  string       `json:"document_name"`
	FilePath      string       `json:"file_path"`
	FileSizeBytes int64        `json:"file_size_bytes"`
	MimeType      string       `json:"mime_type"`
	UploadDate    time.Time    `json:"upload_date"`
}
