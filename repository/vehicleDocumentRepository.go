package repository

import (
	"car_service/database"
	"car_service/entity"
	"context"
)

type VehicleDocumentRepository struct {
}

func NewVehicleDocumentRepository() *VehicleDocumentRepository {
	return &VehicleDocumentRepository{}
}

func (r *VehicleDocumentRepository) InsertVehicleDocument(ctx context.Context, exec database.Executor, document *entity.VehicleDocument) (*entity.VehicleDocument, error) {
	query := `
        INSERT INTO cars.vehicle_documents (vehicle_id, document_type, document_name, file_path, file_size_bytes, mime_type)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, upload_date`

	err := exec.QueryRowContext(ctx, query,
		document.VehicleID,
		document.DocumentType,
		document.DocumentName,
		document.FilePath,
		document.FileSizeBytes,
		document.MimeType,
	).Scan(&document.ID, &document.UploadDate)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (r *VehicleDocumentRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) ([]entity.VehicleDocument, error) {
	query := `
        SELECT id, vehicle_id, document_type, document_name, file_path,
        file_size_bytes, mime_type, upload_date
        FROM cars.vehicle_documents
        WHERE vehicle_id = $1
        ORDER BY upload_date DESC
    `
	rows, err := exec.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []entity.VehicleDocument
	for rows.Next() {
		var doc entity.VehicleDocument
		if err := rows.Scan(
			&doc.ID, &doc.VehicleID, &doc.DocumentType, &doc.DocumentName,
			&doc.FilePath, &doc.FileSizeBytes, &doc.MimeType, &doc.UploadDate,
		); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *VehicleDocumentRepository) GetByID(ctx context.Context, exec database.Executor, id int64) (*entity.VehicleDocument, error) {
	query := `
        SELECT id, vehicle_id, document_type, document_name, file_path,
        file_size_bytes, mime_type, upload_date
        FROM cars.vehicle_documents
        WHERE id = $1
    `

	var doc entity.VehicleDocument
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.VehicleID, &doc.DocumentType, &doc.DocumentName,
		&doc.FilePath, &doc.FileSizeBytes, &doc.MimeType, &doc.UploadDate,
	)

	if err != nil {
		return nil, err
	}

	return &doc, nil
}
