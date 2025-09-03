package main

import (
	"log"
	"net"

	"github.com/go-portfolio/order-pipeline/internal/config" // пакет для загрузки конфигурации приложения
	"github.com/go-portfolio/order-pipeline/internal/server"
	pb "github.com/go-portfolio/order-pipeline/proto" // сгенерированные protobuf файлы
	"github.com/redis/go-redis/v9"                    // клиент для работы с Redis
	"google.golang.org/grpc"                          // gRPC сервер
)

func main() {
	// Загружаем конфигурацию приложения из файла (например, ports, адрес Redis)
	appCfg := config.LoadConfig()

	// Создаём клиент Redis с адресом из конфигурации
	rdb := redis.NewClient(&redis.Options{Addr: appCfg.RedisAddr})

	// Создаём TCP listener на том же адресе
	lis, err := net.Listen("tcp", appCfg.RedisAddr)
	if err != nil {
		log.Fatalf("listen: %v", err) // если порт занят или ошибка сети → завершаем приложение
	}

	// Создаём gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис CacheService в gRPC
	pb.RegisterCacheServiceServer(s, server.NewCacheServer(rdb))

	log.Printf("Сервис кэша слушает на порту %s", appCfg.RedisAddr)

	// Запускаем gRPC сервер и обрабатываем входящие запросы
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err) // если сервер упал → логируем и завершаем
	}
}
