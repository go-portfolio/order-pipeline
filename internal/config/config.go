package config

import (
	"log"
	"os"
)

// Config хранит все переменные окружения проекта
type Config struct {
	KafkaBrokers     string
	KafkaTopic       string
	OrderServiceAddr string
	RedisAddr        string
	CacheServiceAddr string
	DlqTopic         string
	WorkerGroup      string
}

// Load ищет .env вверх от файла и загружает конфигурацию
func Load() *Config {
	cfg := &Config{}
	cfg.KafkaBrokers = os.Getenv("KAFKA_BROKERS")
	cfg.KafkaTopic = os.Getenv("KAFKA_TOPIC")
	cfg.OrderServiceAddr = os.Getenv("ORDER_SERVICE_ADDR")
	cfg.RedisAddr = os.Getenv("REDIS_ADDR")
	cfg.CacheServiceAddr = os.Getenv("CACHE_SERVICE_ADDR")
	cfg.DlqTopic = os.Getenv("DLQ_TOPIC")
	cfg.WorkerGroup = os.Getenv("WORKER_GROUP")

	if cfg.KafkaBrokers == "" || cfg.KafkaTopic == "" || cfg.OrderServiceAddr == "" ||
		cfg.RedisAddr == "" || cfg.CacheServiceAddr == "" || cfg.DlqTopic == "" || cfg.WorkerGroup == "" {
		log.Fatal("Не все переменные окружения для БД установлены")
	}

	return cfg
}

// loadConfig загружает конфигурацию из файла, определяя путь до текущего файла
func LoadConfig() Config {
	return *Load()
}
