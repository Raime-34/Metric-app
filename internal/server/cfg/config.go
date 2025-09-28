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
	Key             string `env:"KEY"`
}

func LoadConfig() {
	env.Parse(&Cfg)

	// 2. зарегистрировали флаги, используя текущие значения как дефолт
	flag.StringVar(&Cfg.Address, "a", Cfg.Address, "Порт на котором будет поднят сервер")
	flag.IntVar(&Cfg.StoreInterval, "i", Cfg.StoreInterval, "Интервал записи метрик в файл")
	flag.StringVar(&Cfg.FileStoragePath, "f", Cfg.FileStoragePath, "Путь к файлу с сохраненными метрика")
	flag.StringVar(&Cfg.DSN, "d", Cfg.DSN, "Параметры подключения к базе данных")
	flag.StringVar(&Cfg.MigrationPath, "m", Cfg.MigrationPath, "Путь к фалам миграции")
	flag.StringVar(&Cfg.Key, "k", Cfg.Key, "Ключ для хэширования")
	flag.BoolVar(&Cfg.Restore, "r", Cfg.Restore, "Флаг для загрузки сохраненных метрик с предыдущего сеанса")

	flag.Parse()
}
