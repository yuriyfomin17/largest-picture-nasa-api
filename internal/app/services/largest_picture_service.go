package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"golang.org/x/sync/errgroup"
	"sort"
	"strconv"
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
func (lps LargestPictureService) PublishCommand(ctx context.Context, sol int) {
	done := make(chan struct{}, 1)
	go func() {
		lps.rabbitMQClient.PublishCommand(sol)
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done(): // Handle context cancellation or timeout
		log.Printf("context cancelled: %v", ctx.Err())
		return
	case <-done: // Return the result of the publish operation
		log.Info().Msg("Command published successfully")
		return
	}
}

func (lps LargestPictureService) GetPictureBySol(ctx context.Context, sol int) (domain.Picture, error) {
	pictureBySol, err := lps.pictureRepo.FindLargestPictureBySol(ctx, sol)

	if err != nil && errors.Is(err, domain.ErrNotFound) {
		return domain.Picture{}, domain.ErrNotFound
	}
	if err != nil && errors.Is(err, domain.ErrCalculationLargestPicture) {
		return domain.Picture{}, domain.ErrCalculationLargestPicture
	}

	if err != nil {
		return domain.Picture{}, err
	}
	return pictureBySol, nil
}

func (lps LargestPictureService) CheckIfPictureExistsSaveIfNecessary(ctx context.Context, sol int) {
	exists, err := lps.pictureRepo.Exists(ctx, sol)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		log.Printf("failed to check if picture exists: %v", err)
		return
	}
	if exists {
		log.Printf("picture for sol %d already exists", sol)
		return
	}
	if err != nil {
		log.Printf("failed to save picture: %v", err)
		return
	}
	lps.findLargestPictureViaAPI(ctx, sol)
}

func (lps LargestPictureService) findLargestPictureViaAPI(ctx context.Context, sol int) {
	photos, err := lps.nasaAPIClient.FindNasaPhotos(ctx, sol)
	if err != nil {
		log.Printf("failed to find nasa photos: %v", err)
		return
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
		log.Printf("errorGroup: %v", errorGroup)
		return
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
		log.Printf("failed to save picture: %v", err)
		return
	}
}

func (lps LargestPictureService) StartListeningSolCommands() {

	go func(ctx context.Context) {
		for message := range lps.rabbitMQClient.GetMessage() {
			strSol := string(message.Body)
			solInt, err := strconv.Atoi(strSol)
			if err != nil {
				log.Printf("failed to convert string to int: %v", err)
				return
			}
			lps.CheckIfPictureExistsSaveIfNecessary(ctx, solInt)
		}

	}(context.TODO())
}
