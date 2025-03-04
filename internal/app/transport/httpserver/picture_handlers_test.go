package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/transport/httpserver/mocks"
)

const commandEndpoint = "/mars/pictures/largest/command"
const getLargestPictureEndpoint = "/mars/pictures/largest/command/{sol}"

func TestHttpServer_PostCommandHandler(t *testing.T) {
	testCases := []struct {
		name                       string
		mockSetup                  func(m *mocks.MarsApiLargestPictureService)
		requestBody                []byte
		expectedStatusCode         int
		expectedResponseBodyShould func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "Should return successful response when request is valid",
			mockSetup: func(m *mocks.MarsApiLargestPictureService) {
				m.On("PublishCommand", mock.Anything, 123).Return(nil)
			},
			requestBody:        []byte(`{"sol": 123}`),
			expectedStatusCode: http.StatusOK,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Equal(t, "Command accepted. Largest picture calculation has started.", body["message"])
			},
		},
		{
			name:               "Should return bad request for invalid JSON body",
			mockSetup:          func(m *mocks.MarsApiLargestPictureService) {}, // No mock calls are needed here
			requestBody:        []byte(`{sol: 123}`),                           // Malformed JSON
			expectedStatusCode: http.StatusBadRequest,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Contains(t, body["slug"], "invalid-command")
			},
		},
		{
			name:               "Should return bad request for missing sol parameter",
			mockSetup:          func(m *mocks.MarsApiLargestPictureService) {}, // No mock calls are needed here
			requestBody:        []byte(`{}`),                                   // Missing "sol"
			expectedStatusCode: http.StatusBadRequest,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Contains(t, body["slug"], "invalid-command")
			},
		},
		{
			name:               "Should return bad request for invalid sol value",
			mockSetup:          func(m *mocks.MarsApiLargestPictureService) {}, // No mock calls are needed here
			requestBody:        []byte(`{"sol": -123}`),                        // Negative value for "sol"
			expectedStatusCode: http.StatusBadRequest,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Contains(t, body["slug"], "invalid-command")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			largestPictureServiceMock := mocks.NewMarsApiLargestPictureService(t)
			tc.mockSetup(largestPictureServiceMock)

			httpServer := NewHttpServer(largestPictureServiceMock)

			req := httptest.NewRequest(http.MethodPost, commandEndpoint, bytes.NewBuffer(tc.requestBody))
			w := httptest.NewRecorder()

			// when
			httpServer.PostCommandHandler(w, req)

			res := w.Result()
			defer res.Body.Close()

			// then
			require.Equal(t, tc.expectedStatusCode, res.StatusCode)

			// Read response body
			var responseBody map[string]interface{}
			err := json.NewDecoder(res.Body).Decode(&responseBody)
			require.NoError(t, err)

			// Assert the response body
			tc.expectedResponseBodyShould(t, responseBody)

			// Verify the mock expectations
			largestPictureServiceMock.AssertExpectations(t)
		})
	}
}

func TestHttpServer_GetLargestPictureHandler(t *testing.T) {
	testCases := []struct {
		name                       string
		sol                        string
		mockSetup                  func(m *mocks.MarsApiLargestPictureService)
		expectedStatusCode         int
		expectedResponseBodyShould func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "Should return largest picture successfully for a valid sol",
			sol:  "123",
			mockSetup: func(m *mocks.MarsApiLargestPictureService) {
				m.On("GetPictureBySol", mock.Anything, 123).Return(domain.NewPicture(domain.NewPictureData{
					Sol:  123,
					Url:  "http://example.com/largest.jpg",
					Size: 1048576,
				}), nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Equal(t, 123.0, body["sol"])
				require.Equal(t, "http://example.com/largest.jpg", body["img_src"])
				require.Equal(t, 1048576.0, body["size"])
				require.Equal(t, "Largest picture fetched successfully", body["message"])
			},
		},
		{
			name:               "Should return bad request when sol is not a valid integer",
			sol:                "abc",
			mockSetup:          func(m *mocks.MarsApiLargestPictureService) {}, // No mock needed
			expectedStatusCode: http.StatusBadRequest,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Contains(t, body["slug"], "invalid-command")
			},
		},
		{
			name: "Should return not found when no picture is found for the given sol",
			sol:  "999",
			mockSetup: func(m *mocks.MarsApiLargestPictureService) {
				m.On("GetPictureBySol", mock.Anything, 999).Return(domain.NewPicture(domain.NewPictureData{}), domain.ErrNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponseBodyShould: func(t *testing.T, body map[string]interface{}) {
				require.Contains(t, body["slug"], "not-found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			largestPictureServiceMock := mocks.NewMarsApiLargestPictureService(t)
			tc.mockSetup(largestPictureServiceMock)

			httpServer := NewHttpServer(largestPictureServiceMock)

			router := mux.NewRouter()
			router.HandleFunc(getLargestPictureEndpoint, httpServer.GetLargestPictureHandler).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/mars/pictures/largest/command/"+tc.sol, nil)
			w := httptest.NewRecorder()

			// when
			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			// then
			require.Equal(t, tc.expectedStatusCode, res.StatusCode)

			// Read response body
			var responseBody map[string]interface{}
			err := json.NewDecoder(res.Body).Decode(&responseBody)
			require.NoError(t, err)

			// Assert the response body
			tc.expectedResponseBodyShould(t, responseBody)

			// Verify the mock expectations
			largestPictureServiceMock.AssertExpectations(t)
		})
	}
}
