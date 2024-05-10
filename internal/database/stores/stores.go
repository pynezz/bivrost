package stores

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/database/models"
	"github.com/pynezz/bivrost/internal/util"
)

type Stores struct {
	NginxLogStore        *database.DataStore[models.NginxLog]
	SynTrafficStore      *database.DataStore[models.SynTraffic]
	AttackTypeStore      *database.DataStore[models.AttackType]
	IndicatorsLogStore   *database.DataStore[models.IndicatorsLog]
	GeoLocationDataStore *database.DataStore[models.GeoLocationData]
	GeoDataStore         *database.DataStore[models.GeoData]

	// Add more stores here if neccessary
	// One store per model (/ table in the/a database)
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
)

var (
	StoreMap map[string]*Stores
)

func new(logDB, moduleDataDB *gorm.DB) (*Stores, error) {
	util.PrintInfo("Initializing stores...")

	util.PrintInfo("Initializing nginx_logs store...")
	nginxLogStore, err := database.NewDataStore[models.NginxLog](logDB, NGINX_LOGS)
	if err != nil {
		return nil, err
	}

	util.PrintInfo("Initializing syn_traffic store with table " + "syn_traffic")
	synTrafficStore, err := database.NewDataStore[models.SynTraffic](logDB, SYN_TRAFFIC)
	if err != nil {
		return nil, err
	}

	indicatorsLogRepo, err := database.NewDataStore[models.IndicatorsLog](moduleDataDB, INDICATORS_LOG)
	if err != nil {
		return nil, err
	}

	geoLocationDataRepo, err := database.NewDataStore[models.GeoLocationData](moduleDataDB, GEO_LOCATION_DATA)
	if err != nil {
		return nil, err
	}

	geoDataRepo, err := database.NewDataStore[models.GeoData](moduleDataDB, GEO_DATA)
	if err != nil {
		return nil, err
	}

	attackTypeRepo, err := database.NewDataStore[models.AttackType](moduleDataDB, ATTACK_TYPE)
	if err != nil {
		return nil, err
	}

	nginxLogStore.Type = models.NginxLog{}
	synTrafficStore.Type = models.SynTraffic{}
	indicatorsLogRepo.Type = models.IndicatorsLog{}
	geoLocationDataRepo.Type = models.GeoLocationData{}
	attackTypeRepo.Type = models.AttackType{}
	geoDataRepo.Type = models.GeoData{}

	util.PrintSuccess("assigned all store types")

	return &Stores{
		NginxLogStore:        nginxLogStore,
		AttackTypeStore:      attackTypeRepo,
		SynTrafficStore:      synTrafficStore,
		IndicatorsLogStore:   indicatorsLogRepo,
		GeoLocationDataStore: geoLocationDataRepo,
		GeoDataStore:         geoDataRepo,
	}, nil
}

func (s *Stores) Get(store string) *Stores {
	util.PrintInfo("Getting store " + store + "...")
	switch store {
	case NGINX_LOGS:
		return &Stores{NginxLogStore: s.NginxLogStore}
	case SYN_TRAFFIC:
		return &Stores{SynTrafficStore: s.SynTrafficStore}
	case INDICATORS_LOG:
		return &Stores{IndicatorsLogStore: s.IndicatorsLogStore}
	case GEO_LOCATION_DATA:
		return &Stores{GeoLocationDataStore: s.GeoLocationDataStore}
	case GEO_DATA:
		return &Stores{GeoDataStore: s.GeoDataStore}
	case ATTACK_TYPE:
		return &Stores{AttackTypeStore: s.AttackTypeStore}
	default:
		return nil
	}
}

// func ok() {
// 	logdb, _ := database.InitDB("logs.db", gorm.Config{}, models.NginxLog{})
// 	modulesdb, _ := database.InitDB("results.db", gorm.Config{}, models.SynTraffic{}, models.AttackType{}, models.IndicatorsLog{}, models.GeoLocationData{}, models.GeoData{})

// 	s, _ := new(logdb, modulesdb)
// 	n := s.Get("nginx_logs")

// 	n.NginxLogStore.InsertLog(models.NginxLog{})

// }

func addToStoreMap(storeName string, store *Stores) {
	StoreMap[storeName] = store
}

func initMap() {
	StoreMap = make(map[string]*Stores)
	util.PrintSuccess("initialized store map")
}

// This is different from the other Get in the way that it's a getter for the StoreMap, rather than an exported method of the Stores struct
func Get(storeName string) *Stores {
	return StoreMap[storeName]
}

func ImportAndInit(conf gorm.Config) (*Stores, error) {
	initMap()
	logdb, _ := database.InitDB("logs.db", conf, models.NginxLog{})
	modulesdb, _ := database.InitDB("results.db", conf, models.SynTraffic{}, models.AttackType{}, models.IndicatorsLog{}, models.GeoLocationData{}, models.GeoData{})

	s, err := new(logdb, modulesdb)
	if err != nil {
		return nil, err
	}

	s.Export()

	util.PrintSuccess("initialized all stores")
	return s, nil
}

func (s *Stores) Export() {

	addToStoreMap("nginx_logs", s.Get(NGINX_LOGS))
	addToStoreMap("syn_traffics", s.Get(SYN_TRAFFIC))
	addToStoreMap("indicators", s.Get(INDICATORS_LOG))
	addToStoreMap("geolocationdata", s.Get(GEO_LOCATION_DATA))
	addToStoreMap("geodata", s.Get(GEO_DATA))
	addToStoreMap("attack_types", s.Get(ATTACK_TYPE))

	util.PrintSuccess("Imported all stores")

}

func Use(store string) (*Stores, error) {
	util.PrintDebug("Using store " + store)
	if ok := StoreMap[store]; ok == nil {
		return nil, fmt.Errorf("store %s not found", store)
	}
	return StoreMap[store], nil
}