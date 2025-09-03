package server

import (
	"context"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)


// orderServer реализует gRPC-сервис OrderService и хранит Kafka writer через интерфейс
type orderServer struct {
	pb.UnimplementedOrderServiceServer
	writer KafkaWriter
}

// NewOrderServer конструктор для инициализации сервера с внедрением зависимости
func NewOrderServer(writer KafkaWriter) pb.OrderServiceServer {
	return &orderServer{writer: writer}
}

// CreateOrder обрабатывает запрос на создание нового заказа
func (s *orderServer) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	msg := kafka.Message{Value: b}

	if err := s.writer.WriteMessages(ctx, msg); err != nil {
		return nil, err
	}

	return &pb.OrderResponse{Status: "accepted"}, nil
}
