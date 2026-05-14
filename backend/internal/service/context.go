package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rillyayidan/nowhere/backend/backend/internal/model"
)

type ContextService struct {
	httpClient *http.Client
}

func NewContextService() *ContextService {
	return &ContextService{
		httpClient: &http.Client{Timeout: 2500 * time.Millisecond},
	}
}

func (s *ContextService) Build(ctx context.Context, req model.DecideRequest) model.Context {
	if req.UserID == "" {
		req.UserID = "demo-user"
	}
	if req.Budget == "" {
		req.Budget = "medium"
	}
	if req.Mood == "" {
		req.Mood = "open"
	}
	if req.TimeOfDay == "" {
		req.TimeOfDay = time.Now().Format("15:04")
	}

	var wg sync.WaitGroup
	weather := fallbackWeather()
	places := fallbackPlaces(req.Location)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if got, err := s.fetchWeather(ctx, req.Location); err == nil {
			weather = got
		}
	}()
	go func() {
		defer wg.Done()
		if got, err := s.fetchPlaces(ctx, req.Location); err == nil && len(got) > 0 {
			places = got
		}
	}()
	wg.Wait()

	return model.Context{
		ID:        uuid.NewString(),
		Location:  req.Location,
		Weather:   weather,
		Places:    places,
		TimeOfDay: req.TimeOfDay,
		Budget:    req.Budget,
		Mood:      req.Mood,
		Rejected:  req.Rejected,
	}
}

func (s *ContextService) fetchWeather(ctx context.Context, loc model.Location) (model.Weather, error) {
	if loc.Latitude == 0 && loc.Longitude == 0 {
		return fallbackWeather(), nil
	}

	endpoint := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,weather_code", loc.Latitude, loc.Longitude)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return model.Weather{}, err
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		return model.Weather{}, err
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		return model.Weather{}, fmt.Errorf("weather status %d", response.StatusCode)
	}

	var payload struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
			Code        int     `json:"weather_code"`
		} `json:"current"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return model.Weather{}, err
	}

	return model.Weather{
		TemperatureC: payload.Current.Temperature,
		Code:         payload.Current.Code,
		Summary:      weatherSummary(payload.Current.Code),
	}, nil
}

func (s *ContextService) fetchPlaces(ctx context.Context, loc model.Location) ([]model.Place, error) {
	key := os.Getenv("GOOGLE_MAPS_KEY")
	if key == "" || (loc.Latitude == 0 && loc.Longitude == 0) {
		return nil, fmt.Errorf("places key or location missing")
	}

	body := map[string]any{
		"includedTypes":  []string{"cafe", "restaurant", "park", "book_store"},
		"maxResultCount": 8,
		"locationRestriction": map[string]any{
			"circle": map[string]any{
				"center": map[string]float64{
					"latitude":  loc.Latitude,
					"longitude": loc.Longitude,
				},
				"radius": 850.0,
			},
		},
	}
	requestPayload, _ := json.Marshal(body)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://places.googleapis.com/v1/places:searchNearby", bytes.NewReader(requestPayload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Goog-Api-Key", key)
	request.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress,places.types,places.primaryType,places.googleMapsUri")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		return nil, fmt.Errorf("places status %d", response.StatusCode)
	}

	var payload struct {
		Results []struct {
			ID          string `json:"id"`
			DisplayName struct {
				Text string `json:"text"`
			} `json:"displayName"`
			Address       string   `json:"formattedAddress"`
			Types         []string `json:"types"`
			PrimaryType   string   `json:"primaryType"`
			GoogleMapsURI string   `json:"googleMapsUri"`
		} `json:"places"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	places := make([]model.Place, 0, len(payload.Results))
	for _, item := range payload.Results {
		category := strings.ReplaceAll(item.PrimaryType, "_", " ")
		if category == "" && len(item.Types) > 0 {
			category = strings.ReplaceAll(item.Types[0], "_", " ")
		}
		if category == "" {
			category = "place"
		}
		mapURL := item.GoogleMapsURI
		if mapURL == "" {
			mapURL = "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(item.DisplayName.Text+" "+item.Address)
		}
		places = append(places, model.Place{
			ID:       item.ID,
			Name:     item.DisplayName.Text,
			Category: category,
			Address:  item.Address,
			OpenNow:  true,
			MapURL:   mapURL,
		})
		if len(places) == 8 {
			break
		}
	}
	return places, nil
}

func fallbackWeather() model.Weather {
	return model.Weather{TemperatureC: 29, Code: 1, Summary: "partly clear"}
}

func fallbackPlaces(loc model.Location) []model.Place {
	return []model.Place{
		{ID: "demo-work-cafe", Name: "Kopi Sela", Category: "cafe", Address: "Near your current area", Rating: 4.7, OpenNow: true, MapURL: mapsURL("Kopi Sela", loc)},
		{ID: "demo-quiet-park", Name: "Taman Lembur", Category: "park", Address: "A short ride away", Rating: 4.5, OpenNow: true, MapURL: mapsURL("Taman Lembur", loc)},
		{ID: "demo-bookshop", Name: "Ruang Baca Kota", Category: "bookstore", Address: "Central neighborhood", Rating: 4.6, OpenNow: true, MapURL: mapsURL("Ruang Baca Kota", loc)},
		{ID: "demo-late-bites", Name: "Warung Nanti", Category: "restaurant", Address: "Main street", Rating: 4.4, OpenNow: true, MapURL: mapsURL("Warung Nanti", loc)},
	}
}

func mapsURL(query string, loc model.Location) string {
	if loc.Latitude == 0 && loc.Longitude == 0 {
		return "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(query)
	}
	return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s&query_place_id=", url.QueryEscape(query))
}

func weatherSummary(code int) string {
	switch {
	case code == 0:
		return "clear"
	case code <= 3:
		return "partly cloudy"
	case code >= 51 && code <= 67:
		return "rainy"
	case code >= 80 && code <= 82:
		return "showers"
	case code >= 95:
		return "stormy"
	default:
		return "mild"
	}
}
