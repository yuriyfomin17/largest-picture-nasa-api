package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"golang.org/x/sync/errgroup"
)

type LargestPictureService struct {
	rabbitMQClient RabbitMQClient
	pictureRepo    PictureRepository
	nasaAPIClient  NasaAPIClient
}

func NewLargestPictureService(
	rabbitMQClient RabbitMQClient,
	pictureRepo PictureRepository,
	nasaApiClient NasaAPIClient,
) LargestPictureService {
	return LargestPictureService{
		rabbitMQClient: rabbitMQClient,
		pictureRepo:    pictureRepo,
		nasaAPIClient:  nasaApiClient,
	}
}
func (lps LargestPictureService) PublishCommand(ctx context.Context, sol int) error {
	return lps.rabbitMQClient.PublishCommand(ctx, sol)
}

func (lps LargestPictureService) GetPictureBySol(ctx context.Context, sol int) (domain.Picture, error) {
	pictureBySol, err := lps.pictureRepo.FindLargestPictureBySol(ctx, sol)

	if err != nil && errors.Is(err, domain.ErrNotFound) {
		return domain.Picture{}, domain.ErrNotFound
	}

	if err != nil {
		return domain.Picture{}, domain.ErrCalculationLargestPicture
	}
	return pictureBySol, nil
}

func (lps LargestPictureService) CheckIfPictureExistsSaveIfNecessary(ctx context.Context, sol int) error {
	exists, err := lps.pictureRepo.Exists(ctx, sol)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("failed to check if picture exists: %w", err)
	}
	if exists {
		return fmt.Errorf("picture for sol %d already exists", sol)
	}
	if err != nil {
		return fmt.Errorf("failed to save picture: %w", err)
	}
	err = lps.findLargestPictureViaAPI(ctx, sol)
	if err != nil {
		return fmt.Errorf("failed to find largest picture via API: %w", err)
	}
	return nil
}

func (lps LargestPictureService) findLargestPictureViaAPI(ctx context.Context, sol int) error {
	photos, err := lps.nasaAPIClient.FindNasaPhotos(ctx, sol)
	if err != nil {
		return fmt.Errorf("failed to find nasa photos: %w", err)
	}
	g, currContext := errgroup.WithContext(ctx)
	g.SetLimit(len(photos.Photos))
	nasaPhotoChannels := make(chan domain.NewPictureData, len(photos.Photos))

	for _, photo := range photos.Photos {
		g.Go(func() error {
			size, currError := lps.nasaAPIClient.FindPhotoSize(&currContext, photo.ImageSrc)
			if currError != nil {
				return fmt.Errorf("findLargestPicture %w", currError)
			}
			currNasaPicture := domain.NewPictureData{
				Size: size,
				Sol:  sol,
				Url:  photo.ImageSrc,
			}
			select {
			case <-currContext.Done():
				fmt.Println("error was called so context is done")
				return currContext.Err()
			case nasaPhotoChannels <- currNasaPicture:
				return nil
			}

		})
	}
	if errorGroup := g.Wait(); errorGroup != nil {
		return fmt.Errorf("errorGroup: %w", errorGroup)
	}
	close(nasaPhotoChannels)
	nasaPhotos := make([]domain.NewPictureData, len(photos.Photos))
	for nasaPhoto := range nasaPhotoChannels {
		nasaPhotos = append(nasaPhotos, nasaPhoto)
	}
	sort.Slice(nasaPhotos, func(i, j int) bool {
		return nasaPhotos[i].Size >= nasaPhotos[j].Size
	})
	err = lps.pictureRepo.Save(ctx, domain.NewPicture(nasaPhotos[0]))
	if err != nil {
		return fmt.Errorf("failed to save picture: %w", err)
	}
	return nil
}

func (lps LargestPictureService) StartListeningSolCommands(ctx context.Context) {
	go func() {
		for message := range lps.rabbitMQClient.GetMessage() {
			strSol := string(message.Body)
			solInt, err := strconv.Atoi(strSol)
			if err != nil {
				log.Printf("failed to convert string to int: %v", err)
				return
			}
			err = lps.CheckIfPictureExistsSaveIfNecessary(ctx, solInt)
			if err != nil {
				log.Printf("failed to check if picture exists: %v", err)
				return
			}
		}
	}()
}
