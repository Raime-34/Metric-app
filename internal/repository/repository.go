package repository

import (
	models "metricapp/internal/model"
)

type Repo interface {
	SetField(string, models.Metrics)
	GetFields() map[string]models.Metrics
	IncrementCounter()
}
