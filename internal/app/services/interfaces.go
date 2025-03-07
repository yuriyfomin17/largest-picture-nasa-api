//go:generate mockery

package services

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/clients/models"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
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

type RabbitMQClient interface {
	PublishCommand(ctx context.Context, solCommand int) error
	GetMessage() <-chan amqp.Delivery
}
