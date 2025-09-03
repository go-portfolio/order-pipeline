package server

import (
	"context"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// cacheServer реализует gRPC-сервис CacheService и хранит подключение к Redis
type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	rdb *redis.Client // клиент для работы с Redis
}

// NewCacheServer конструктор для инициализации сервера
func NewCacheServer(rdb *redis.Client) pb.CacheServiceServer {
	return &cacheServer{rdb: rdb}
}

// GetOrderResult обрабатывает запрос на получение результата заказа по ID
func (s *cacheServer) GetOrderResult(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	// формируем ключ для Redis
	key := "order:" + req.Id

	// пытаемся получить значение из Redis
	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// ключ не найден → возвращаем gRPC ошибку NotFound
		return nil, status.Error(codes.NotFound, "order not found")
	} else if err != nil {
		// любая другая ошибка Redis → Internal
		return nil, status.Error(codes.Internal, "redis error: "+err.Error())
	}

	// десериализуем JSON из Redis в protobuf-структуру
	var res pb.ResultResponse
	if err := protojson.Unmarshal([]byte(val), &res); err != nil {
		// ошибка при разборе JSON → Internal
		return nil, status.Error(codes.Internal, "unmarshal error: "+err.Error())
	}

	// возвращаем результат
	return &res, nil
}
