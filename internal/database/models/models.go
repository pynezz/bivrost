package models

func GetModels() []interface{} {
	return []interface{}{
		&NginxLog{},
		&SynTraffic{},
		&AttackType{},
		&IndicatorsLog{},
		&GeoLocationData{},
		&GeoData{},

		// We'll add more models here if neccessary
	}
}

const (
	NGINX_LOGS        = "nginx_logs"
	SYN_TRAFFIC       = "syntraffic"
	ATTACK_TYPE       = "attack_type"
	INDICATORS_LOG    = "indicators"
	GEO_LOCATION_DATA = "geodata_details"
	GEO_DATA          = "geodata"

	// We'll add more model names here as well if neccessary. This is useful for being able to reference the model names with . notation
	// (e.g. models.NGINX_LOGS, models.SYN_TRAFFIC, etc. providing autocomplete support)
)
