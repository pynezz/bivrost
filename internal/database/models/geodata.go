package models

import "gorm.io/gorm"

type GeoLocationData struct {
	gorm.Model
	Query       string
	Country     string
	CountryCode string
	RegionName  string
	Lat         float64
	Lon         float64
	Message     string
}

type GeoData struct {
	gorm.Model
	Source            string
	GeoLocationDataID uint
	GeoLocationData   GeoLocationData `gorm:"foreignKey:GeoLocationDataID"`
}

type GeoDataModel interface {
	GeoData         // Embed the GeoData interface
	GeoLocationData // Embed the GeoLocationData interface

	GetGeoData() *GeoData
}

func (g *GeoData) Store() *GeoData {
	return g
}
