package models

type Workload struct {
	Model

	ID                   uint    `gorm:"primary_key;not null"`
	UUID                 string  `gorm:"type:varchar(36);not null"`
	TaskUUID             string  `gorm:"type:varchar(36);not null"`
	SubtaskUUID          string  `gorm:"type:varchar(36);not null"`
	Name                 string  `gorm:"type:text"`
	Description          string  `gorm:"type:text"`
	Position             float64 `gorm:"type:float"`
	Runner               string  `gorm:"type:text"`
	RunnerType           string  `gorm:"type:text"`
	Contexts             string  `gorm:"type:text"`
	ContextsResults      string  `gorm:"type:text"`
	Sla                  string  `gorm:"type:text"`
	SlaResults           string  `gorm:"type:text"`
	Args                 string  `gorm:"type:text"`
	Hooks                string  `gorm:"type:text"`
	StartTime            float64 `gorm:"type:float"`
	LoadDuration         float64 `gorm:"type:float"`
	FullDuration         float64 `gorm:"type:float"`
	MinDuration          float64 `gorm:"type:float"`
	MaxDuration          float64 `gorm:"type:float"`
	TotalIterationCount  float64 `gorm:"type:float"`
	FailedIterationCount float64 `gorm:"type:float"`
	Statistics           string  `gorm:"type:text"`
	PassSLA              bool
}

type StatisticsFormat struct {
	Durations struct {
		Total struct {
			Data struct {
				IterationCount int     `json:"iteration_count"`
				Min            float64 `json:"min"`
				Median         float64 `json:"median"`
				Nine0Ile       float64 `json:"90%ile"`
				Nine5Ile       float64 `json:"95%ile"`
				Max            float64 `json:"max"`
				Avg            float64 `json:"avg"`
				Success        string  `json:"success"`
			} `json:"data"`
			CountPerIteration int    `json:"count_per_iteration"`
			Name              string `json:"name"`
			DisplayName       string `json:"display_name"`
			Children          []struct {
				Data struct {
					IterationCount int     `json:"iteration_count"`
					Min            float64 `json:"min"`
					Median         float64 `json:"median"`
					Nine0Ile       float64 `json:"90%ile"`
					Nine5Ile       float64 `json:"95%ile"`
					Max            float64 `json:"max"`
					Avg            float64 `json:"avg"`
					Success        string  `json:"success"`
				} `json:"data"`
				CountPerIteration int    `json:"count_per_iteration"`
				Name              string `json:"name"`
				DisplayName       string `json:"display_name"`
				Children          []any  `json:"children"`
			} `json:"children"`
		} `json:"total"`
		Atomics []struct {
			Data struct {
				IterationCount int     `json:"iteration_count"`
				Min            float64 `json:"min"`
				Median         float64 `json:"median"`
				Nine0Ile       float64 `json:"90%ile"`
				Nine5Ile       float64 `json:"95%ile"`
				Max            float64 `json:"max"`
				Avg            float64 `json:"avg"`
				Success        string  `json:"success"`
			} `json:"data"`
			CountPerIteration int    `json:"count_per_iteration"`
			Name              string `json:"name"`
			DisplayName       string `json:"display_name"`
			Children          []any  `json:"children"`
		} `json:"atomics"`
	} `json:"durations"`
}
