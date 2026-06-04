package agent

import "time"

type BriefRequest struct {
	Issuer   string `json:"issuer"`
	Question string `json:"question"`
	Live     bool   `json:"live,omitempty"`
}

type Evidence struct {
	Source      string `json:"source"`
	Title       string `json:"title"`
	Observation string `json:"observation"`
	URL         string `json:"url,omitempty"`
}

type SpecialistResult struct {
	Agent      string     `json:"agent"`
	Summary    string     `json:"summary"`
	Confidence string     `json:"confidence"`
	Evidence   []Evidence `json:"evidence"`
}

type BriefResponse struct {
	Issuer      string             `json:"issuer"`
	Question    string             `json:"question"`
	GeneratedAt time.Time          `json:"generated_at"`
	Mode        string             `json:"mode"`
	Plan        []string           `json:"plan"`
	Findings    []SpecialistResult `json:"findings"`
	Brief       string             `json:"brief"`
	NextAction  string             `json:"next_action"`
	Disclosures []string           `json:"disclosures"`
}
