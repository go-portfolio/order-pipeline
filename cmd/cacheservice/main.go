package main

import (
	"context"
	"log"
	"net"

	"github.com/go-portfolio/order-pipeline/internal/config"
	"github.com/go-portfolio/order-pipeline/internal/server"
	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	// Загружаем конфигурацию приложения
	appCfg := config.LoadConfig()

	// Подключаемся к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: appCfg.RedisAddr, // Redis: redis:6379
	})

	// Проверим подключение к Redis (ping с контекстом)
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("cannot connect to Redis at %s: %v", appCfg.RedisAddr, err)
	}

	// Создаём TCP listener для gRPC сервера на отдельном порту
	lis, err := net.Listen("tcp", appCfg.CacheServiceAddr) // gRPC: 0.0.0.0:50052
	if err != nil {
		log.Fatalf("cannot listen on %s: %v", appCfg.CacheServiceAddr, err)
	}

	// Создаём gRPC сервер
	s := grpc.NewServer()

	// Регистрируем сервис CacheService
	pb.RegisterCacheServiceServer(s, server.NewCacheServer(rdb))

	log.Printf("CacheService listening on %s", appCfg.CacheServiceAddr)

	// Запускаем gRPC сервер
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC serve failed: %v", err)
	}
}
