//go:generate mockery

package httpserver

import (
	"context"
	"largest-picture-nasa-api/internal/app/domain"
)

// MarsApiLargestPictureService is picture service
type MarsApiLargestPictureService interface {
	GetPictureBySol(ctx context.Context, sol int) (domain.Picture, error)
	CheckIfPictureExistsSaveIfNecessary(ctx context.Context, sol int)
	PublishCommand(ctx context.Context, sol int)
	StartListeningSolCommands()
}
