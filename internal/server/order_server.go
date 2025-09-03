package server

import (
	"context"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

// server реализует gRPC-сервис OrderService и хранит Kafka writer
type orderServer struct {
	pb.UnimplementedOrderServiceServer
	kafkaWriter *kafka.Writer // writer для публикации сообщений в Kafka
}

// NewCacheServer конструктор для инициализации сервера
func NewOrderServer(kafkaWriter *kafka.Writer) pb.OrderServiceServer {
	return &orderServer{kafkaWriter: kafkaWriter}
}

// CreateOrder обрабатывает запрос на создание нового заказа
func (s *orderServer) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	// сериализуем protobuf-сообщение в байты
	b, err := proto.Marshal(req)
	if err != nil {
		// если сериализация не удалась → возвращаем ошибку
		return nil, err
	}

	// создаём сообщение Kafka
	msg := kafka.Message{Value: b}

	// публикуем сообщение в Kafka
	if err := s.kafkaWriter.WriteMessages(ctx, msg); err != nil {
		// если ошибка записи в Kafka → возвращаем ошибку
		return nil, err
	}

	// возвращаем успешный статус клиенту
	return &pb.OrderResponse{Status: "accepted"}, nil
}
