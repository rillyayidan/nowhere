package model

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy,omitempty"`
	Label     string  `json:"label,omitempty"`
}

type DecideRequest struct {
	UserID        string   `json:"user_id"`
	Location      Location `json:"location"`
	Budget        string   `json:"budget"`
	Mood          string   `json:"mood"`
	TimeOfDay     string   `json:"time_of_day"`
	Rejected      []string `json:"rejected"`
	LastContextID string   `json:"last_context_id,omitempty"`
}

type FeedbackRequest struct {
	UserID     string   `json:"user_id"`
	ContextID  string   `json:"context_id"`
	DecisionID string   `json:"decision_id"`
	Action     string   `json:"action"`
	Place      Place    `json:"place"`
	Rejected   []string `json:"rejected,omitempty"`
}
