package order

import (
	"context"

	"github.com/segmentio/kafka-go"                    // Библиотека для работы с Kafka
	pb "github.com/go-portfolio/order-pipeline/proto" // gRPC-сгенерированные структуры и интерфейсы
	"google.golang.org/protobuf/proto"                // Для сериализации protobuf-сообщений
)

// Server реализует интерфейс gRPC OrderService и содержит Kafka writer
type Server struct {
	pb.UnimplementedOrderServiceServer // Встраиваемая структура с "заглушками" для gRPC-методов
	KafkaWriter *kafka.Writer          // Объект для записи сообщений в Kafka
}

// CreateOrder обрабатывает gRPC-запрос на создание заказа
func (s *Server) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	// Сериализуем protobuf-запрос в байты
	b, err := proto.Marshal(req)
	if err != nil {
		// Возвращаем ошибку, если сериализация не удалась
		return nil, err
	}

	// Создаём Kafka-сообщение с сериализованным телом
	msg := kafka.Message{Value: b}

	// Пытаемся отправить сообщение в Kafka
	if err := s.KafkaWriter.WriteMessages(ctx, msg); err != nil {
		// Возвращаем ошибку, если Kafka недоступна или произошёл сбой
		return nil, err
	}

	// Возвращаем успешный ответ клиенту
	return &pb.OrderResponse{Status: "accepted"}, nil
}
