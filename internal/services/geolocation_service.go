package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"delivery-system/internal/config"
	"delivery-system/internal/logger"
	"delivery-system/internal/models"
	"delivery-system/internal/redis"

	"golang.org/x/net/context"
)

const (
	defaultCacheTTL = 15 * time.Hour
)

/*
GeolocationService - сервис для работы с геосервисами.
GeolocationService.yandexKey - API-ключ для работы с Яндекс-Геокодер для получения координат адресов
GeolocationService.openrouteKey - API-ключ для работы с OpenrouteService для расчёта маршрута между точками
*/
type GeolocationService struct {
	openrouteKey string
	yandexKey    string
	redisClient  *redis.Client
	log          *logger.Logger
}

// NewGeolocationService создаёт новый экземпляр геосервиса
func NewGeolocationService(cfg *config.GeolocationConfig, redisClient *redis.Client, log *logger.Logger) *GeolocationService {
	return &GeolocationService{
		openrouteKey: cfg.OperouteAPIKey,
		yandexKey:    cfg.YandexAPIKey,
		redisClient:  redisClient,
		log:          log,
	}
}

// GetCoordinates возвращает координаты (lng, lat) указанного адреса
func (g *GeolocationService) GetCoordinates(address string) (float64, float64, error) {
	escapedAddress := url.QueryEscape(address)

	requestURL := fmt.Sprintf(
		"%s?api_key=%s&geocode=%s&results=%d&format=%s",
		models.YandexGeocoderURL,
		g.yandexKey,
		escapedAddress,
		1,
		"json",
	)

	resp, err := http.Get(requestURL)
	if err != nil {
		g.log.WithError(err).Error("Failed to get response from Yandex API")
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		g.log.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"respBody":    string(body),
		}).Error("Bad response from Yandex API")
		return 0, 0, fmt.Errorf("bad response with status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		g.log.Error("Failed to read response body")
		return 0, 0, err
	}

	var apiResponse models.YandexResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		g.log.WithError(err).Error("Failed to unmarshal Yandex API response")
		return 0, 0, err
	}

	featureMember := apiResponse.Response.GeoObjectCollection.FeatureMember
	if len(featureMember) == 0 {
		g.log.Warn("No objects in response body")
		return 0, 0, fmt.Errorf("couldn't get coordinates from address: %s", address)
	}

	pos := featureMember[0].GeoObject.Point.Pos
	var lng, lat float64
	_, err = fmt.Sscanf(pos, "%f %f", &lng, &lat)
	if err != nil {
		g.log.WithError(err).Error("Failed to parse coordinates")
		return 0, 0, fmt.Errorf("failed to parse coordinates: %w", err)
	}

	return lng, lat, nil
}

// MakeRoute возвращает длину маршрута, построенного по переданным координатам
func (g *GeolocationService) MakeRoute(coordinates [][2]float64) (float64, error) {
	requestBody := models.OpenrouteRequest{
		Coordinates:  coordinates,
		Instructions: "false",
		Maneuvers:    "false",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		g.log.WithError(err).Error("Failed to marshal request body")
		return 0, err
	}

	req, err := http.NewRequest("POST", models.OpenrouteDirectionsURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		g.log.WithError(err).Error("Failed to create request")
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", g.openrouteKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		g.log.WithError(err).Error("Failed to send request")
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		g.log.WithFields(map[string]interface{}{
			"status_code": resp.StatusCode,
			"respBody":    string(body),
		}).Error("Bad response from Openroute API")
		return 0, fmt.Errorf("bad response with status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		g.log.WithError(err).Error("Failed to read response body")
		return 0, err
	}

	var apiResponse models.OpenrouteResponse
	if err := json.Unmarshal(data, &apiResponse); err != nil {
		g.log.WithError(err).Error("Failed to unmarshal Openroute API response")
	}

	dist := apiResponse.Routes.Summary.Distance
	return dist, nil
}

func (g *GeolocationService) cacheResults(coordinates [][2]float64, distance float64, order *models.Order) {
	orderGeolocation := models.GeoCache{
		PickupCoordinates:   coordinates[0],
		DeliveryCoordinates: coordinates[1],
		Distance:            distance,
		DeliveryCost:        order.DeliveryCost,
	}
	cacheKey := redis.GenerateKey(redis.KeyPrefixOrderGeolocation, order.ID.String())
	if err := g.redisClient.Set(context.Background(), cacheKey, orderGeolocation, defaultCacheTTL); err != nil {
		g.log.WithError(err).Error("Failed to cache order")
	}
}
