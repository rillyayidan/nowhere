package model

type Weather struct {
	TemperatureC float64 `json:"temperature_c"`
	Code         int     `json:"code"`
	Summary      string  `json:"summary"`
}

type Place struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Address  string  `json:"address"`
	Rating   float64 `json:"rating,omitempty"`
	OpenNow  bool    `json:"open_now"`
	MapURL   string  `json:"map_url,omitempty"`
}

type Context struct {
	ID        string   `json:"id"`
	Location  Location `json:"location"`
	Weather   Weather  `json:"weather"`
	Places    []Place  `json:"places"`
	TimeOfDay string   `json:"time_of_day"`
	Budget    string   `json:"budget"`
	Mood      string   `json:"mood"`
	Rejected  []string `json:"rejected"`
}

type Decision struct {
	ID              string `json:"id"`
	Place           Place  `json:"place"`
	Reason          string `json:"reason"`
	DurationMinutes int    `json:"duration_minutes"`
	Confidence      string `json:"confidence"`
}

type DecideResponse struct {
	Context  Context  `json:"context"`
	Decision Decision `json:"decision"`
	Source   string   `json:"source"`
	Error    string   `json:"error,omitempty"`
}

type FeedbackResponse struct {
	Status         string `json:"status"`
	AcceptedCount  int    `json:"accepted_count"`
	RejectedCount  int    `json:"rejected_count"`
	PreferenceHint string `json:"preference_hint"`
}
