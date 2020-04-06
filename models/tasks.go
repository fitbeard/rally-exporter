package models

type Task struct {
	Model

	ID                 uint    `gorm:"primary_key;not null"`
	UUID               string  `gorm:"type:varchar(36);not null"`
	EnvUUID            string  `gorm:"type:varchar(36);not null"`
	InputTask          string  `gorm:"type:text"`
	Title              string  `gorm:"type:varchar(128)"`
	Description        string  `gorm:"type:text"`
	ValidationResult   string  `gorm:"type:text;not null"`
	ValidationDuration float64 `gorm:"type:float"`
	TaskDuration       float64 `gorm:"type:float"`
	PassSLA            bool
	Status             string `gorm:"type:varchar(36)"`
}

type Subtask struct {
	Model

	ID              uint    `gorm:"primary_key;not null"`
	UUID            string  `gorm:"type:varchar(36);not null"`
	TaskUUID        string  `gorm:"type:varchar(36);not null"`
	Title           string  `gorm:"type:varchar(128)"`
	Description     string  `gorm:"type:text"`
	Contexts        string  `gorm:"type:text;not null"`
	ContextsResults string  `gorm:"type:text;not null"`
	SLA             string  `gorm:"type:text;not null"`
	RunInParallel   bool    `gorm:"not null"`
	Duration        float64 `gorm:"type:float"`
	PassSLA         bool
	Status          string `gorm:"type:varchar(36)"`
}
