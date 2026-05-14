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

type PreferenceStore struct {
	mu         sync.RWMutex
	httpClient *http.Client
	accepted   map[string][]model.Place
	rejected   map[string][]model.Place
}

func NewPreferenceStore() *PreferenceStore {
	return &PreferenceStore{
		httpClient: &http.Client{Timeout: 2500 * time.Millisecond},
		accepted:   map[string][]model.Place{},
		rejected:   map[string][]model.Place{},
	}
}

func (s *PreferenceStore) Save(ctx context.Context, req model.FeedbackRequest) model.FeedbackResponse {
	userID := req.UserID
	if userID == "" {
		userID = "demo-user"
	}

	s.mu.Lock()

	action := strings.ToLower(req.Action)
	if action == "accept" {
		s.accepted[userID] = append(s.accepted[userID], req.Place)
	} else {
		s.rejected[userID] = append(s.rejected[userID], req.Place)
	}
	hint := s.HintLocked(userID)
	acceptedCount := len(s.accepted[userID])
	rejectedCount := len(s.rejected[userID])
	s.mu.Unlock()

	status := "stored-local"
	if err := s.writeFirestore(ctx, userID, req); err == nil {
		status = "stored"
	}
	return model.FeedbackResponse{
		Status:         status,
		AcceptedCount:  acceptedCount,
		RejectedCount:  rejectedCount,
		PreferenceHint: hint,
	}
}

func (s *PreferenceStore) Hint(userID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.HintLocked(userID)
}

func (s *PreferenceStore) HintLocked(userID string) string {
	if userID == "" {
		userID = "demo-user"
	}
	if len(s.accepted[userID]) == 0 && len(s.rejected[userID]) == 0 {
		return "No stored feedback yet. Use the first decision as a baseline."
	}
	if len(s.accepted[userID]) >= len(s.rejected[userID]) {
		return "Lean toward places similar to accepted choices and keep the explanation direct."
	}
	return "Avoid recently rejected categories and pick a clearly different option."
}

func (s *PreferenceStore) writeFirestore(ctx context.Context, userID string, req model.FeedbackRequest) error {
	projectID := firstEnv("FIREBASE_PROJECT_ID", "GOOGLE_CLOUD_PROJECT", "GOOGLE_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("firestore project missing")
	}

	token, err := s.cloudAccessToken(ctx)
	if err != nil {
		return err
	}

	documentID := uuid.NewString()
	endpoint := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents/users/%s/feedback?documentId=%s", projectID, url.PathEscape(userID), documentID)
	payload := firestoreDocument(map[string]any{
		"user_id":     userID,
		"context_id":  req.ContextID,
		"decision_id": req.DecisionID,
		"action":      req.Action,
		"place_id":    req.Place.ID,
		"place_name":  req.Place.Name,
		"category":    req.Place.Category,
		"address":     req.Place.Address,
		"rating":      req.Place.Rating,
		"created_at":  time.Now().UTC().Format(time.RFC3339),
	})
	body, _ := json.Marshal(payload)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode > 299 {
		return fmt.Errorf("firestore status %d", response.StatusCode)
	}
	return nil
}

func (s *PreferenceStore) cloudAccessToken(ctx context.Context) (string, error) {
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
		return "", err
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

func firestoreDocument(values map[string]any) map[string]any {
	fields := map[string]any{}
	for key, value := range values {
		switch typed := value.(type) {
		case string:
			fields[key] = map[string]string{"stringValue": typed}
		case float64:
			fields[key] = map[string]float64{"doubleValue": typed}
		case bool:
			fields[key] = map[string]bool{"booleanValue": typed}
		}
	}
	return map[string]any{"fields": fields}
}
