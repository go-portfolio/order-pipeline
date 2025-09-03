package main

import (
	"context"
	"log"
	"net"

	"github.com/go-portfolio/order-pipeline/internal/config" // пакет для загрузки конфигурации приложения
	pb "github.com/go-portfolio/order-pipeline/proto"        // сгенерированные protobuf файлы
	"github.com/redis/go-redis/v9"                           // клиент для работы с Redis
	"google.golang.org/grpc"                                 // gRPC сервер
	"google.golang.org/grpc/codes"                           // коды ошибок gRPC
	"google.golang.org/grpc/status"                          // формирование ошибок gRPC
	"google.golang.org/protobuf/encoding/protojson"          // кодирование/декодирование protobuf в JSON
)

// cacheServer реализует gRPC-сервис CacheService и хранит подключение к Redis
type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	rdb *redis.Client // клиент для работы с Redis
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
	pb.RegisterCacheServiceServer(s, &cacheServer{rdb: rdb})

	log.Printf("Сервис кэша слушает на порту %s", appCfg.RedisAddr)

	// Запускаем gRPC сервер и обрабатываем входящие запросы
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err) // если сервер упал → логируем и завершаем
	}
}
