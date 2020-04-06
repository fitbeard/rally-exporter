package models

type Environment struct {
	Model

	ID          uint   `gorm:"primary_key;not null"`
	UUID        string `gorm:"type:varchar(36);not null"`
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
	Status      string `gorm:"type:varchar(36);not null"`
	Extras      string `gorm:"type:text"`
	Config      string `gorm:"type:text"`
	Spec        string `gorm:"type:text"`
}
