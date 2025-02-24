package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/repository/models"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/pkg"
)

type PictureRepo struct {
	db *pkg.DB
}

func NewPictureRepo(db *pkg.DB) PictureRepo {
	return PictureRepo{db: db}
}

// FindLargestPictureBySol retrieves the largest picture (based on size) for the given sol
func (r *PictureRepo) FindLargestPictureBySol(ctx context.Context, sol int) (domain.Picture, error) {
	var picture models.Picture

	err := r.db.NewSelect().Model(&picture).
		Where("sol = ?", sol).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Picture{}, domain.ErrNotFound // No rows found, so it doesn't exist
	}
	if err != nil {
		return domain.Picture{}, fmt.Errorf("could not find largest picture by sol: %w", err)
	}

	return toDomainPicture(picture), nil
}

// Save inserts or updates a picture record
func (r *PictureRepo) Save(ctx context.Context, picture domain.Picture) error {
	modelPicture := domainToPicture(picture)
	_, err := r.db.NewInsert().
		Model(&modelPicture).
		On("CONFLICT (sol) DO UPDATE").
		Set("img_src = EXCLUDED.img_src, size = EXCLUDED.size"). // Update specific fields in case of conflict
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not save picture: %w", err)
	}
	return nil
}

// Exists checks if a picture exists in the database for the given sol
func (r *PictureRepo) Exists(ctx context.Context, sol int) (bool, error) {
	var exists bool

	err := r.db.NewSelect().
		Model((*domain.Picture)(nil)). // Use the domain.Picture model for the table
		ColumnExpr("1").               // Doesn't fetch full data, just checks for existence
		Where("sol = ?", sol).
		Limit(1). // We only need to check if one exists
		Scan(ctx, &exists)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not check if picture exists: %w", err)
	}

	return exists, nil
}
