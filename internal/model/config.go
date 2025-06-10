package model

import "time"

type AppConfig struct {
	Key       string    `gorm:"primaryKey;size:100"`
	Value     string    `gorm:"type:text;not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
