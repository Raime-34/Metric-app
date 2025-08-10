package cfg

import (
	"flag"

	"github.com/caarlos0/env"
)

var Cfg Config

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DSN             string `env:"DATABASE_DSN"`
	MigrationPath   string `env:"MIGRATION_PATH"`
}

func LoadConfig() {
	env.Parse(&Cfg)

	if Cfg.Address == "" {
		flag.StringVar(&Cfg.Address, "a", "0.0.0.0:8080", "Порт на котором будет поднят сервер")
	}
	if Cfg.StoreInterval == 0 {
		flag.IntVar(&Cfg.StoreInterval, "i", 300, "Интервал записи метрик в файл")
	}
	if Cfg.FileStoragePath == "" {
		flag.StringVar(&Cfg.FileStoragePath, "f", "./metrics.log", "Путь к файлу с сохраненными метрика")
	}
	if Cfg.DSN == "" {
		flag.StringVar(&Cfg.DSN, "d", "", "Параметры подключения к базе даннных")
	}
	if Cfg.MigrationPath == "" {
		flag.StringVar(&Cfg.MigrationPath, "m", "migrations", "Путь к фалам миграции")
	}
	var restore bool
	flag.BoolVar(&restore, "r", false, "Флаг для загрузки сохраненных метрик с предыдущего сеанса")
	if !Cfg.Restore {
		Cfg.Restore = restore
	}
	flag.Parse()
}
