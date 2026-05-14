package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rillyayidan/nowhere/backend/backend/internal/model"
)

type DecisionService struct {
	httpClient *http.Client
	prefs      *PreferenceStore
}

func NewDecisionService(prefs *PreferenceStore) *DecisionService {
	return &DecisionService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		prefs:      prefs,
	}
}

func (s *DecisionService) Decide(ctx context.Context, contextData model.Context, userID string) (model.Decision, string, error) {
	if decision, source, err := s.askGemini(ctx, contextData, userID); err == nil && decision.Place.Name != "" {
		decision.ID = uuid.NewString()
		return decision, source, nil
	} else if !allowDemoFallback() {
		return model.Decision{}, "", err
	}
	return s.fallbackDecision(contextData), "demo-fallback", nil
}

func (s *DecisionService) askGemini(ctx context.Context, contextData model.Context, userID string) (model.Decision, string, error) {
	if useVertexAI() {
		return s.askVertex(ctx, contextData, userID)
	}
	return s.askDeveloperAPI(ctx, contextData, userID)
}

func (s *DecisionService) askDeveloperAPI(ctx context.Context, contextData model.Context, userID string) (model.Decision, string, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return model.Decision{}, "", fmt.Errorf("GEMINI_API_KEY missing")
	}

	payload, _ := json.Marshal(geminiRequest(s.prompt(contextData, userID)))
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", geminiModel(), key)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return model.Decision{}, "", err
	}
	request.Header.Set("Content-Type", "application/json")
	decision, err := s.sendGeminiRequest(request, contextData)
	return decision, "gemini-developer-api", err
}

func (s *DecisionService) askVertex(ctx context.Context, contextData model.Context, userID string) (model.Decision, string, error) {
	projectID := firstEnv("GOOGLE_CLOUD_PROJECT", "GOOGLE_PROJECT_ID", "GCLOUD_PROJECT")
	location := firstEnv("GOOGLE_CLOUD_LOCATION", "GOOGLE_CLOUD_REGION", "VERTEX_AI_LOCATION")
	if projectID == "" || location == "" {
		return model.Decision{}, "", fmt.Errorf("GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION are required for Vertex AI")
	}

	token, err := s.vertexAccessToken(ctx)
	if err != nil {
		return model.Decision{}, "", err
	}

	payload, _ := json.Marshal(geminiRequest(s.prompt(contextData, userID)))
	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent", location, projectID, location, geminiModel())
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return model.Decision{}, "", err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	decision, err := s.sendGeminiRequest(request, contextData)
	return decision, "vertex-ai", err
}

func (s *DecisionService) sendGeminiRequest(request *http.Request, contextData model.Context) (model.Decision, error) {
	response, err := s.httpClient.Do(request)
	if err != nil {
		return model.Decision{}, err
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1200))
		return model.Decision{}, fmt.Errorf("gemini status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var gemini struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(response.Body).Decode(&gemini); err != nil {
		return model.Decision{}, err
	}
	if len(gemini.Candidates) == 0 || len(gemini.Candidates[0].Content.Parts) == 0 {
		return model.Decision{}, fmt.Errorf("empty gemini response")
	}

	var raw struct {
		PlaceName       string `json:"place"`
		PlaceID         string `json:"place_id"`
		Reason          string `json:"reason"`
		DurationMinutes int    `json:"duration_minutes"`
		Confidence      string `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(gemini.Candidates[0].Content.Parts[0].Text), &raw); err != nil {
		return model.Decision{}, err
	}

	place := findPlace(contextData.Places, raw.PlaceID, raw.PlaceName)
	if place.Name == "" {
		place = model.Place{ID: raw.PlaceID, Name: raw.PlaceName, Category: "place", OpenNow: true}
	}
	return model.Decision{
		Place:           place,
		Reason:          raw.Reason,
		DurationMinutes: raw.DurationMinutes,
		Confidence:      raw.Confidence,
	}, nil
}

func (s *DecisionService) vertexAccessToken(ctx context.Context) (string, error) {
	if token := os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN"); token != "" {
		return token, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Metadata-Flavor", "Google")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("unable to get Cloud Run service account token: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		return "", fmt.Errorf("metadata token status %d", response.StatusCode)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("metadata server returned an empty token")
	}
	return payload.AccessToken, nil
}

func (s *DecisionService) prompt(contextData model.Context, userID string) string {
	places, _ := json.Marshal(contextData.Places)
	rejected := strings.Join(contextData.Rejected, ", ")
	if rejected == "" {
		rejected = "none"
	}
	return fmt.Sprintf(`You are NowHere, a context-aware place decision engine.
Choose exactly one place from this JSON list and respond only with JSON.

Context:
- weather: %.1f C, %s
- time: %s
- budget: %s
- mood: %s
- preference hint: %s
- excluded place ids or names: %s

Places JSON:
%s

Required JSON shape:
{"place":"name","place_id":"id","reason":"short reason under 24 words","duration_minutes":90,"confidence":"high|medium|low"}`,
		contextData.Weather.TemperatureC,
		contextData.Weather.Summary,
		contextData.TimeOfDay,
		contextData.Budget,
		contextData.Mood,
		s.prefs.Hint(userID),
		rejected,
		string(places),
	)
}

func (s *DecisionService) fallbackDecision(contextData model.Context) model.Decision {
	rejected := map[string]bool{}
	for _, item := range contextData.Rejected {
		rejected[strings.ToLower(item)] = true
	}

	chosen := model.Place{}
	for _, place := range contextData.Places {
		if rejected[strings.ToLower(place.ID)] || rejected[strings.ToLower(place.Name)] {
			continue
		}
		if chosen.Name == "" || scorePlace(place, contextData) > scorePlace(chosen, contextData) {
			chosen = place
		}
	}
	if chosen.Name == "" && len(contextData.Places) > 0 {
		chosen = contextData.Places[0]
	}

	reason := fmt.Sprintf("%s fits the %s weather, your %s budget, and keeps the next step simple.", chosen.Name, contextData.Weather.Summary, contextData.Budget)
	return model.Decision{
		ID:              uuid.NewString(),
		Place:           chosen,
		Reason:          reason,
		DurationMinutes: 75,
		Confidence:      "medium",
	}
}

func scorePlace(place model.Place, contextData model.Context) float64 {
	score := place.Rating
	if place.OpenNow {
		score += 0.5
	}
	category := strings.ToLower(place.Category)
	if strings.Contains(contextData.Weather.Summary, "rain") && (strings.Contains(category, "cafe") || strings.Contains(category, "restaurant")) {
		score += 1
	}
	if contextData.Budget == "low" && strings.Contains(category, "park") {
		score += 1
	}
	if strings.Contains(strings.ToLower(contextData.Mood), "quiet") && (strings.Contains(category, "book") || strings.Contains(category, "park")) {
		score += 1
	}
	return score
}

func findPlace(places []model.Place, id string, name string) model.Place {
	id = strings.ToLower(id)
	name = strings.ToLower(name)
	for _, place := range places {
		if strings.ToLower(place.ID) == id || strings.ToLower(place.Name) == name {
			return place
		}
	}
	return model.Place{}
}

func geminiRequest(prompt string) map[string]any {
	return map[string]any{
		"contents": []map[string]any{{
			"role":  "user",
			"parts": []map[string]string{{"text": prompt}},
		}},
		"generationConfig": map[string]any{
			"temperature":      0.35,
			"maxOutputTokens":  256,
			"responseMimeType": "application/json",
		},
	}
}

func geminiModel() string {
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		return "gemini-2.5-flash-lite"
	}
	return modelName
}

func useVertexAI() bool {
	value := strings.ToLower(os.Getenv("GOOGLE_GENAI_USE_VERTEXAI"))
	return value == "1" || value == "true" || value == "yes"
}

func allowDemoFallback() bool {
	value := strings.ToLower(os.Getenv("NOWHERE_ALLOW_DEMO_FALLBACK"))
	return value == "" || value == "1" || value == "true" || value == "yes"
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
