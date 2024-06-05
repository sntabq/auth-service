package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	// Env — текущее виртуальное окружение.
	Env string `yaml:"env" env-default:"local"`
	// StoragePath — путь до файла, где хранится база данных.
	StoragePath string `yaml:"storage_path" env-required:"true"`
	// GRPCConﬁg — порт gRPC-сервиса и таймаут обработки запросов.
	GRPC GRPCConfig `yaml:"grpc"`
	// MigrationsPath — путь до директории с миграциями базы данных, который будет использовать утилита migrator.
	MigrationsPath string
	// TokenTTL — время жизни выдаваемых токенов авторизации. Для простоты сделаем фиксированным и будем хранить в конфигурации.
	TokenTTL time.Duration `yaml:"token_ttl" env-default:"1h"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	// --config=path/to/file
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
