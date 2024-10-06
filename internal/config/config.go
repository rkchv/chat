package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const localConfigPath = "./.env"

// Config Конфиги grpc сервера, бд и прочего. Можно задавать через env, можно в yml конфиге
type Config struct {
	Env string `yaml:"env" env:"ENV" env-required:"true"`
	GRPC
	Postgres
	SecretKey        string        `yaml:"secret_key" env:"JWT_SECRET_KEY" env-required:"true"`
	AuthServiceAddr  string        `yaml:"auth_service_addr" env:"AUTH_SERVICE_ADDR" env-default:"localhost:50501"`
	ChatGarbageCycle time.Duration `yaml:"chat_garbage_cycle" env:"CHAT_GARBAGE_CYCLE" env-default:"10s"`
	ChatExpired      time.Duration `yaml:"chat_expired" env:"CHAT_EXPIRED" env-default:"1m"`
	Trace
	Prometheus
}

// MustLoad загружает конфиг из окружения/файла. Фаталится если не получится
func MustLoad() Config {

	var cfg Config

	errEnv := cleanenv.ReadEnv(&cfg)
	if errEnv == nil {
		return cfg
	}

	//если из окружения не получили нужные параметры, пробуем взять конфиг файл
	cfgPath := os.Getenv("CONFIG_PATH")

	if cfgPath == "" {
		if _, err := os.Stat(localConfigPath); os.IsNotExist(err) {
			log.Fatalf("config path not set and env reading error: %v", errEnv)
		}

		cfgPath = localConfigPath
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Fatalf("config file not exists: %s", cfgPath)
	}

	err := cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}

	return cfg
}
