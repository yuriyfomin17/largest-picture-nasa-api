package services

import (
	"context"
	"largest-picture-nasa-api/internal/app/clients/models"
	"largest-picture-nasa-api/internal/app/domain"
)

type PictureRepository interface {
	FindLargestPictureBySol(ctx context.Context, sol int) (domain.Picture, error)
	Save(ctx context.Context, picture domain.Picture) error
	Exists(ctx context.Context, sol int) (bool, error)
}

type NasaAPIClient interface {
	FindNasaPhotos(ctx context.Context, sol int) (models.NasaPhotos, error)
	FindPhotoSize(ctx *context.Context, imgUrl string) (int, error)
}
