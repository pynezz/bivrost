package fetcher

// More info: https://gosamples.dev/sqlite-intro/

type NginxLog struct {
	ID            int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
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

type NginxLogsList struct {
	Logs []NginxLog
}

type NginxLogsRepository interface {
	Migrate() error
	Create(log NginxLog) (*NginxLog, error)
	All() ([]NginxLog, error)
	GetByIP(name string) (*NginxLog, error)
	Update(id int64, updated NginxLog) (*NginxLog, error)
	Delete(id int64) error
}
