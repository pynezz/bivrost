package models

import "gorm.io/gorm"

type NginxLog struct {
	gorm.Model // Includes fields ID, CreatedAt, UpdatedAt, DeletedAt

	ID            int64  `json:"-" sqlite:"-"` // Unique identifier - autoincremented, so no need to set it
	TimeLocal     string `json:"time_local"`
	RemoteAddr    string `json:"remote_addr"`
	RemoteUser    string `json:"remote_user"`
	Request       string `json:"request"`
	Status        string `json:"status"`
	BodyBytesSent string `json:"body_bytes_sent"`
	RequestTime   string `json:"request_time"`
	HttpReferer   string `json:"http_referrer"`
	HttpUserAgent string `json:"http_user_agent"`
	RequestBody   string `json:"request_body"`
}
