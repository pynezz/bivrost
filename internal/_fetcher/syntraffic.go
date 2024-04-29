package fetcher

type SynTrafficDetectionLog struct {
	ID             int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Description    string `json:"description"`
	Source         string `json:"source"`
	FirstTimestamp string `json:"first_timestamp"`
	LastTimestamp  string `json:"last_timestamp"`
	Count          int    `json:"count"`
	Status         string `json:"status"`
	Recommendation string `json:"recommendation"`
}
