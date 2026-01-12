package repository

import (
	"car_service/database"
	"car_service/entity"
	"context"
	"fmt"
)

type VehicleImageRepository struct {
}

func NewVehicleImageRepository() *VehicleImageRepository {
	return &VehicleImageRepository{}
}

func (s *VehicleImageRepository) InsertVehicleImage(ctx context.Context, exec database.Executor, vehicleImage *entity.VehicleImage) (*entity.VehicleImage, error) {
	query := `
        INSERT INTO cars.vehicle_images (vehicle_id, filename, original_name, file_path, file_size, mime_type, is_primary, display_order)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, upload_date`

	err := exec.QueryRowContext(ctx, query,
		vehicleImage.VehicleID,
		vehicleImage.Filename,
		vehicleImage.OriginalName,
		vehicleImage.FilePath,
		vehicleImage.FileSize,
		vehicleImage.MimeType,
		vehicleImage.IsPrimary,
		vehicleImage.DisplayOrder,
	).Scan(&vehicleImage.ID, &vehicleImage.UploadDate)

	if err != nil {
		return nil, err
	}

	return vehicleImage, nil
}

func (s *VehicleImageRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) ([]entity.VehicleImage, error) {
	query := `
        SELECT id, vehicle_id, filename, original_name, file_path,
        file_size, mime_type, is_primary, upload_date, display_order
        FROM cars.vehicle_images
        WHERE vehicle_id = $1
        ORDER BY display_order ASC
    `
	rows, err := exec.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []entity.VehicleImage
	for rows.Next() {
		var img entity.VehicleImage
		if err := rows.Scan(
			&img.ID, &img.VehicleID, &img.Filename, &img.OriginalName,
			&img.FilePath, &img.FileSize, &img.MimeType, &img.IsPrimary,
			&img.UploadDate, &img.DisplayOrder,
		); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

// SetPrimaryImage sets a specific image as primary and unsets all others for the same vehicle
func (s *VehicleImageRepository) SetPrimaryImage(ctx context.Context, exec database.Executor, imageID int, vehicleID int64) error {
	// First, unset all primary images for this vehicle
	unsetQuery := `
		UPDATE cars.vehicle_images
		SET is_primary = false
		WHERE vehicle_id = $1
	`
	_, err := exec.ExecContext(ctx, unsetQuery, vehicleID)
	if err != nil {
		return err
	}

	// Then set the specified image as primary
	setPrimaryQuery := `
		UPDATE cars.vehicle_images
		SET is_primary = true
		WHERE id = $1 AND vehicle_id = $2
	`
	result, err := exec.ExecContext(ctx, setPrimaryQuery, imageID, vehicleID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with id %d not found for vehicle %d", imageID, vehicleID)
	}

	return nil
}
