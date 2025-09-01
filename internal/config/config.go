package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// Config хранит все переменные окружения проекта
type Config struct {
	KafkaBrokers     string
	KafkaTopic       string
	OrderServiceAddr string
	RedisAddr        string
	CacheServiceAddr string
}

// Load ищет .env вверх от файла и загружает конфигурацию
func Load(file string) *Config {
	dir := filepath.Dir(file)
	envPath := ""
	found := false

	for {
		envPath = filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				log.Fatal("Ошибка загрузки .env файла:", err)
			}
			found = true
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if !found {
		log.Fatal(".env файл не найден")
	}

	cfg := &Config{}
	cfg.KafkaBrokers = os.Getenv("KAFKA_BROKERS")
	cfg.KafkaTopic = os.Getenv("KAFKA_TOPIC")
	cfg.OrderServiceAddr = os.Getenv("ORDER_SERVICE_ADDR")
	cfg.RedisAddr = os.Getenv("REDIS_ADDR")
	cfg.CacheServiceAddr = os.Getenv("CACHE_SERVICE_ADDR")

	if cfg.KafkaBrokers == "" || cfg.KafkaTopic == "" || cfg.OrderServiceAddr == "" {
		log.Fatal("Не все переменные окружения для БД установлены")
	}

	return cfg
}

// loadConfig загружает конфигурацию из файла, определяя путь до текущего файла
func LoadConfig() Config {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("не удалось определить путь до конфигурации")
	}
	return *Load(filename)
}
