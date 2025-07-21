package filemanager

import (
	"encoding/json"
	"io"
	models "metricapp/internal/model"
	"os"
)

type FManager struct {
	file          os.File
	Storeinterval int
}

func Open(path string, sInterval int) (*FManager, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FManager{file: *file, Storeinterval: sInterval}, nil
}

func (fm *FManager) Write(metrics []models.Metrics) error {
	b, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	fm.file.Truncate(0)
	fm.file.Seek(0, 0)
	_, err = fm.file.Write(b)
	return err
}

func (fm *FManager) Read() ([]models.Metrics, error) {
	b, err := io.ReadAll(&fm.file)

	if err != nil {
		return nil, err
	}

	var metrics []models.Metrics
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (fm *FManager) Close() error {
	return fm.file.Close()
}
