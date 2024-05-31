package models

import "gorm.io/gorm"

type ThreatRecord struct {
	gorm.Model

	ID          int64  `json:"-"` // Unique identifier - autoincremented, so no need to set it
	Description string `yaml:"description"`
	Source      string `yaml:"source"`
	Timestamp   string `yaml:"timestamp"`
	Status      string `yaml:"status"`
	Severity    string `yaml:"severity"`
	File        string `yaml:"file"` // New field
}
