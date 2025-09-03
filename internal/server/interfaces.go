package server

import (
	"context"
	"time"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Интерфейсы для тестирования
type KafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

// cacheServer реализует gRPC-сервис CacheService и хранит подключение к Redis через интерфейс
type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	rdb RedisClient
}
