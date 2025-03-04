package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sheepla/go-urlbuilder"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/clients/models"
)

var HTTPClient = http.Client{
	Timeout: time.Second * 5,
}

type NasaApiClient struct {
	apiKey string
	apiUrl *urlbuilder.URL
}

func NewNasaApiClient(apiKey string, apiUrl string) NasaApiClient {
	parsedUrl := urlbuilder.MustParse(apiUrl)
	return NasaApiClient{
		apiKey: apiKey,
		apiUrl: parsedUrl,
	}
}

func (c NasaApiClient) FindNasaPhotos(ctx context.Context, sol int) (models.NasaPhotos, error) {
	solStr := strconv.Itoa(sol)
	var currentUrl = c.buildUrl(c.apiKey, solStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, currentUrl, nil)
	if err != nil {
		return models.NasaPhotos{}, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return models.NasaPhotos{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.NasaPhotos{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return models.NasaPhotos{}, fmt.Errorf("unexpected status code: %d, response body: %s\n", resp.StatusCode, responseBytes)
	}

	var photos models.NasaPhotos
	if err := json.Unmarshal(responseBytes, &photos); err != nil {
		return models.NasaPhotos{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	if len(photos.Photos) == 0 {
		return models.NasaPhotos{}, fmt.Errorf("no photo found")
	}
	return photos, nil
}

func (c NasaApiClient) FindPhotoSize(ctx *context.Context, imgUrl string) (int, error) {
	req, err := http.NewRequestWithContext(*ctx, "HEAD", imgUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("could not send request: %w", err)
	}
	return int(resp.ContentLength), nil
}

func (c NasaApiClient) buildUrl(apiKey, sol string) string {
	c.apiUrl.EditQuery(func(q url.Values) url.Values {
		q.Set("sol", sol)
		q.Set("api_key", apiKey)
		return q
	})
	return c.apiUrl.MustString()
}
