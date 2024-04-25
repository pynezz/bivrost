package models

type IndicatorsLog struct {
	ID                  int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Source              string `json:"ip"`
	AbuseIPDBMalicious  bool   `json:"abuseipdb_blacklisted"`
	AlienVaultMalicious bool   `json:"alien_vault_blacklisted"`
	ThreatFoxMalicious  bool   `json:"threat_fox_blacklisted"`
	Timestamp           string `json:"timestamp"`
}
