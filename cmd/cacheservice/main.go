package main

import (
	"context"
	"log"
	"net"

	"github.com/go-portfolio/order-pipeline/internal/config"
	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type cacheServer struct {
	pb.UnimplementedCacheServiceServer
	rdb *redis.Client
}

func (s *cacheServer) GetOrderResult(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	key := "order:" + req.Id
	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, status.Error(codes.NotFound, "order not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "redis error: "+err.Error())
	}
	var res pb.ResultResponse
	if err := protojson.Unmarshal([]byte(val), &res); err != nil {
		return nil, status.Error(codes.Internal, "unmarshal error: "+err.Error())
	}
	return &res, nil
}

func main() {
	// Загружаем конфигурацию приложения из файла
	appCfg := config.LoadConfig()

	rdb := redis.NewClient(&redis.Options{Addr: appCfg.RedisAddr})

	lis, err := net.Listen("tcp", appCfg.RedisAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCacheServiceServer(s, &cacheServer{rdb: rdb})
	log.Printf("CacheService listening on %s", appCfg.RedisAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
