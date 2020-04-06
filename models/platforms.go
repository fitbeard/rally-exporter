package models

type Platform struct {
	Model

	ID           uint   `gorm:"primary_key;not null"`
	UUID         string `gorm:"type:varchar(36);not null"`
	EnvUUID      string `gorm:"type:varchar(36);not null"`
	Status       string `gorm:"type:varchar(36);not null"`
	PluginName   string `gorm:"type:varchar(36);not null"`
	PluginSpec   string `gorm:"type:text;not null"`
	PluginData   string `gorm:"type:text"`
	PlatformName string `gorm:"type:varchar(36)"`
	PlatformData string `gorm:"type:text"`
}
