package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/clients/models"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/services/mocks"
)

func TestLargestPictureService_PublishCommand(t *testing.T) {
	testCases := []struct {
		name      string
		sol       int
		mockSetup func(m *mocks.RabbitMqclient)
	}{
		{
			name: "Should publish command successfully",
			sol:  123,
			mockSetup: func(m *mocks.RabbitMqclient) {
				m.On("PublishCommand", 123).Return(nil).Once()
			},
		},
		{
			name: "Should handle context cancellation before publishing",
			sol:  456,
			mockSetup: func(m *mocks.RabbitMqclient) {
				// Simulate no call since context cancels
				m.On("PublishCommand", mock.Anything).Return(nil).Maybe()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			mockRabbitMQ := mocks.NewRabbitMqclient(t)
			tc.mockSetup(mockRabbitMQ)
			lps := NewLargestPictureService(mockRabbitMQ, nil, nil)

			ctx, cancel := context.WithCancel(context.Background())
			if tc.name == "Should handle context cancellation before publishing" {
				cancel() // Trigger context cancellation
			}

			// When
			lps.PublishCommand(ctx, tc.sol)

			// Then
			mockRabbitMQ.AssertExpectations(t)
		})
	}
}

func TestLargestPictureService_GetPictureBySol(t *testing.T) {
	testCases := []struct {
		name           string
		sol            int
		mockSetup      func(m *mocks.PictureRepository)
		expectedError  error
		expectedResult interface{}
	}{
		{
			name: "Should return a picture successfully",
			sol:  123,
			mockSetup: func(m *mocks.PictureRepository) {
				m.On("FindLargestPictureBySol", mock.Anything, 123).Return(
					domain.NewPicture(domain.NewPictureData{Sol: 123, Url: "http://example.com", Size: 1048576}),
					nil,
				).Once()
			},
			expectedError: nil,
			expectedResult: domain.NewPicture(domain.NewPictureData{
				Sol:  123,
				Url:  "http://example.com",
				Size: 1048576,
			}),
		},
		{
			name: "Should return not found error if picture doesn't exist",
			sol:  456,
			mockSetup: func(m *mocks.PictureRepository) {
				m.On("FindLargestPictureBySol", mock.Anything, 456).Return(domain.Picture{}, domain.ErrNotFound).Once()
			},
			expectedError:  domain.ErrNotFound,
			expectedResult: domain.Picture{},
		},
		{
			name: "Should return unexpected error on repository failure",
			sol:  789,
			mockSetup: func(m *mocks.PictureRepository) {
				m.On("FindLargestPictureBySol", mock.Anything, 789).Return(domain.Picture{}, errors.New("unexpected error")).Once()
			},
			expectedError:  errors.New("unexpected error"),
			expectedResult: domain.Picture{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			mockRepo := mocks.NewPictureRepository(t)
			tc.mockSetup(mockRepo)
			lps := NewLargestPictureService(nil, mockRepo, nil)

			// When
			result, err := lps.GetPictureBySol(context.Background(), tc.sol)

			// Then
			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLargestPictureService_CheckIfPictureExistsSaveIfNecessary(t *testing.T) {
	testCases := []struct {
		name      string
		sol       int
		mockSetup func(repo *mocks.PictureRepository, apiClient *mocks.NasaApiclient)
	}{
		{
			name: "Should not save picture if it already exists",
			sol:  123,
			mockSetup: func(repo *mocks.PictureRepository, apiClient *mocks.NasaApiclient) {
				repo.On("Exists", mock.Anything, 123).Return(true, nil).Once()
			},
		},
		{
			name: "Should handle unexpected repository error",
			sol:  999,
			mockSetup: func(repo *mocks.PictureRepository, apiClient *mocks.NasaApiclient) {
				repo.On("Exists", mock.Anything, 999).Return(false, errors.New("repository error")).Once()
			},
		},
		{
			name: "Should call Nasa API and save the largest picture when it doesn't exist",
			sol:  456,
			mockSetup: func(repo *mocks.PictureRepository, apiClient *mocks.NasaApiclient) {
				repo.On("Exists", mock.Anything, 456).Return(false, nil).Once()
				apiClient.On("FindNasaPhotos", mock.Anything, 456).Return(models.NasaPhotos{
					Photos: []models.NasaPhoto{
						{ImageSrc: "http://example1.com"},
						{ImageSrc: "http://example2.com"},
					},
				}, nil).Once()
				apiClient.On("FindPhotoSize", mock.AnythingOfType("*context.Context"), "http://example1.com").Return(1024, nil).Once()
				apiClient.On("FindPhotoSize", mock.AnythingOfType("*context.Context"), "http://example2.com").Return(2048, nil).Once()

				repo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			mockRepo := mocks.NewPictureRepository(t)
			mockApi := mocks.NewNasaApiclient(t)
			tc.mockSetup(mockRepo, mockApi)

			lps := NewLargestPictureService(nil, mockRepo, mockApi)

			// When
			lps.CheckIfPictureExistsSaveIfNecessary(context.Background(), tc.sol)

			// Then
			mockRepo.AssertExpectations(t)
			mockApi.AssertExpectations(t)
		})
	}
}
