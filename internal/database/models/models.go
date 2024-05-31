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

/*
results.db

attack_types       geo_location_data  syn_traffics
geo_data           indicators_logs
*/
const (
	NGINX_LOGS        = "nginx_logs"
	SYN_TRAFFIC       = "syn_traffics"
	ATTACK_TYPE       = "attack_types"
	INDICATORS_LOG    = "indicators_logs"
	GEO_LOCATION_DATA = "geo_location_data"
	GEO_DATA          = "geo_data"
	THREAT_RECORDS    = "threat_records"

	// We'll add more model names here as well if neccessary. This is useful for being able to reference the model names with . notation
	// (e.g. models.NGINX_LOGS, models.SYN_TRAFFIC, etc. providing autocomplete support)
)
