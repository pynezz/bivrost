package stores

import (
	"gorm.io/gorm"

	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/database/models"
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

var (
	StoreMap map[string]*Stores
)

func new(logDB, moduleDataDB *gorm.DB) (*Stores, error) {
	nginxLogStore, err := database.NewDataStore[models.NginxLog](logDB, "nginx_logs")
	if err != nil {
		return nil, err
	}

	synTrafficStore, err := database.NewDataStore[models.SynTraffic](logDB, "syn_traffic")
	if err != nil {
		return nil, err
	}

	indicatorsLogRepo, err := database.NewDataStore[models.IndicatorsLog](moduleDataDB, "indicatorslog")
	if err != nil {
		return nil, err
	}

	geoLocationDataRepo, err := database.NewDataStore[models.GeoLocationData](moduleDataDB, "geolocationdata")
	if err != nil {
		return nil, err
	}

	geoDataRepo, err := database.NewDataStore[models.GeoData](moduleDataDB, "geodata")
	if err != nil {
		return nil, err
	}

	return &Stores{
		NginxLogStore:        nginxLogStore,
		SynTrafficStore:      synTrafficStore,
		IndicatorsLogStore:   indicatorsLogRepo,
		GeoLocationDataStore: geoLocationDataRepo,
		GeoDataStore:         geoDataRepo,
	}, nil
}

func (s *Stores) Get(store string) *Stores {
	switch store {
	case "nginx_logs":
		return &Stores{NginxLogStore: s.NginxLogStore}
	case "syn_traffic":
		return &Stores{SynTrafficStore: s.SynTrafficStore}
	case "indicators":
		return &Stores{IndicatorsLogStore: s.IndicatorsLogStore}
	case "geolocationdata":
		return &Stores{GeoLocationDataStore: s.GeoLocationDataStore}
	case "geodata":
		return &Stores{GeoDataStore: s.GeoDataStore}
	default:
		return nil
	}
}

func ok() {
	logdb, _ := database.InitDB("logs.db", gorm.Config{}, models.NginxLog{})
	modulesdb, _ := database.InitDB("results.db", gorm.Config{}, models.SynTraffic{}, models.AttackType{}, models.IndicatorsLog{}, models.GeoLocationData{}, models.GeoData{})

	s, _ := new(logdb, modulesdb)
	n := s.Get("nginx_logs")

	n.NginxLogStore.InsertLog(models.NginxLog{})

}

func addToStoreMap(storeName string, store *Stores) {
	StoreMap[storeName] = store
}

func initMap() {
	StoreMap = make(map[string]*Stores)
}

// This is different from the other Get in the way that it's a getter for the StoreMap, rather than an exported method of the Stores struct
func Get(storeName string) *Stores {
	return StoreMap[storeName]
}

func Import() (*Stores, error) {
	initMap()
	logdb, _ := database.InitDB("logs.db", gorm.Config{}, models.NginxLog{})
	modulesdb, _ := database.InitDB("results.db", gorm.Config{}, models.SynTraffic{}, models.AttackType{}, models.IndicatorsLog{}, models.GeoLocationData{}, models.GeoData{})

	s, err := new(logdb, modulesdb)
	if err != nil {
		return nil, err
	}

	return s, nil
}
