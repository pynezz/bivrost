package fetcher

// type AttackTypeLog struct {
// 	ID          int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
// 	Source      string `json:"remote_addr"`
// 	Description string `json:"description"`
// 	TimeLocal   string `json:"time_local"`
// 	UserAgent   string `json:"user_agent"`
// 	Payload     string `json:"payload"`
// }

type AttackTypeLog struct {
	ID             int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Source         string `json:"source"`
	Description    string `json:"description"`
	Count          int    `json:"count"`
	Severity       string `json:"severity"`
	Threshold      string `json:"threshold"`
	FirstTimestamp string `json:"first_timestamp"`
	LastTimestamp  string `json:"last_timestamp"`
	Status         string `json:"status"`
	Recommendation string `json:"recommendation"`
	RequestPath    string `json:"request_path"`
	UserAgent      string `json:"user_agent"`
	Payload        string `json:"payload"`
}
