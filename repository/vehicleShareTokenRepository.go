package repository

import (
	"car_service/database"
	"car_service/entity"
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type VehicleShareTokenRepository struct{}

func NewVehicleShareTokenRepository() *VehicleShareTokenRepository {
	return &VehicleShareTokenRepository{}
}

// Insert creates a new vehicle share token
func (r *VehicleShareTokenRepository) Insert(ctx context.Context, exec database.Executor, token *entity.VehicleShareToken) (int64, error) {
	query := `
		INSERT INTO cars.vehicle_share_tokens
		(vehicle_id, token, expires_at, include_details, created_by, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := exec.QueryRowContext(
		ctx,
		query,
		token.VehicleID,
		token.Token,
		token.ExpiresAt,
		pq.Array(token.IncludeDetails),
		token.CreatedBy,
		true,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetByToken retrieves a share token by its token string
func (r *VehicleShareTokenRepository) GetByToken(ctx context.Context, exec database.Executor, token string) (*entity.VehicleShareToken, error) {
	query := `
		SELECT
			id,
			vehicle_id,
			token,
			expires_at,
			include_details,
			created_by,
			created_at,
			is_active
		FROM cars.vehicle_share_tokens
		WHERE token = $1 AND is_active = true AND expires_at > NOW()
	`

	var shareToken entity.VehicleShareToken
	err := exec.QueryRowContext(ctx, query, token).Scan(
		&shareToken.ID,
		&shareToken.VehicleID,
		&shareToken.Token,
		&shareToken.ExpiresAt,
		pq.Array(&shareToken.IncludeDetails),
		&shareToken.CreatedBy,
		&shareToken.CreatedAt,
		&shareToken.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &shareToken, nil
}

// GetByVehicleID retrieves all active tokens for a vehicle
func (r *VehicleShareTokenRepository) GetByVehicleID(ctx context.Context, exec database.Executor, vehicleID int64) ([]entity.VehicleShareToken, error) {
	query := `
		SELECT
			id,
			vehicle_id,
			token,
			expires_at,
			include_details,
			created_by,
			created_at,
			is_active
		FROM cars.vehicle_share_tokens
		WHERE vehicle_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := exec.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []entity.VehicleShareToken
	for rows.Next() {
		var token entity.VehicleShareToken
		err := rows.Scan(
			&token.ID,
			&token.VehicleID,
			&token.Token,
			&token.ExpiresAt,
			pq.Array(&token.IncludeDetails),
			&token.CreatedBy,
			&token.CreatedAt,
			&token.IsActive,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

// DeactivateToken deactivates a share token
func (r *VehicleShareTokenRepository) DeactivateToken(ctx context.Context, exec database.Executor, tokenID int64) error {
	query := `
		UPDATE cars.vehicle_share_tokens
		SET is_active = false
		WHERE id = $1
	`

	_, err := exec.ExecContext(ctx, query, tokenID)
	return err
}

// DeactivateExpiredTokens deactivates all expired tokens
func (r *VehicleShareTokenRepository) DeactivateExpiredTokens(ctx context.Context, exec database.Executor) error {
	query := `
		UPDATE cars.vehicle_share_tokens
		SET is_active = false
		WHERE expires_at < NOW() AND is_active = true
	`

	_, err := exec.ExecContext(ctx, query)
	return err
}
