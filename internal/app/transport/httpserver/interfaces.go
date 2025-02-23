//go:generate mockery

package httpserver

import (
	"context"
	"largest-picture-nasa-api/internal/app/domain"
)

// MarsAPILargestPictureService is picture service
type MarsAPILargestPictureService interface {
	GetPictureBySol(ctx context.Context, sol int) (domain.Picture, error)
	CheckIfPictureExistsSaveIfNecessary(ctx context.Context, sol int)
	PublishCommand(ctx context.Context, sol int)
	StartListeningSolCommands()
}
