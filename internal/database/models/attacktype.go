package models

import "gorm.io/gorm"

type AttackType struct {
	gorm.Model

	ID             int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Description    string `json:"description"`
	Source         string `json:"source"`
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
