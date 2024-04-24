package fetcher

type ThreatTypeLog struct {
	ID          int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Source      string `json:"remote_addr"`
	Description string `json:"description"`
	TimeLocal   string `json:"time_local"`
	UserAgent   string `json:"user_agent"`
	Payload     string `json:"payload"`
}
