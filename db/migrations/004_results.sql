/* sqlite3 database */
-- Path: db/migrations/004_results.sql

-- for detecting nmap scans
CREATE TABLE IF NOT EXISTS syntraffic (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT,
    description TEXT,
    first_timestamp TIMESTAMP,
    last_timestamp TIMESTAMP,
    count INTEGER,
    status TEXT,
    recommendation TEXT
);

CREATE TABLE IF NOT EXISTS attack_type (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT,
    description TEXT,
    count INTEGER,
    severity TEXT,
    threshold TEXT,
    first_timestamp TIMESTAMP,
    last_timestamp TIMESTAMP,
    status TEXT,
    recommendation TEXT,
    request_path TEXT,
    user_agent TEXT,
    payload TEXT
);

/*
type IndicatorsLog struct {
	ID                  int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Source              string `json:"ip"`
	AbuseIPDBMalicious  bool   `json:"abuseipdb_blacklisted"`
	AlienVaultMalicious bool   `json:"alien_vault_blacklisted"`
	ThreatFoxMalicious  bool   `json:"threat_fox_blacklisted"`
	Timestamp           string `json:"timestamp"`
}
*/
CREATE TABLE IF NOT EXISTS indicators (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT,
    abuseipdb_blacklisted BOOLEAN,
    alien_vault_blacklisted BOOLEAN,
    threat_fox_blacklisted BOOLEAN,
    timestamp TIMESTAMP
);

/* // Main display for Geolocation Json File
type GeoLocationData struct {
	ID          int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Query       string  `json:"query"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	RegionName  string  `json:"regionName"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Message     string  `json:"message"`
} */

CREATE TABLE IF NOT EXISTS geodata_details (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query TEXT,
    country TEXT,
    countryCode TEXT,
    regionName TEXT,
    lat REAL,
    lon REAL,
    message TEXT
);

/* type GeoLocationDataWrapper struct {
	Source  string          `json:"source"`
	GeoData GeoLocationData `json:"geolocation_data`
} */

CREATE TABLE IF NOT EXISTS geodata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT,
    geolocation_data_id INTEGER,
    FOREIGN KEY (geolocation_data_id) REFERENCES geodata(id)
);
