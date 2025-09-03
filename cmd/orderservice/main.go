package main

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/go-portfolio/order-pipeline/internal/config" // пакет для загрузки конфигурации приложения
	pb "github.com/go-portfolio/order-pipeline/proto"        // сгенерированные protobuf файлы для OrderService
	"github.com/segmentio/kafka-go"                           // клиент Kafka для записи сообщений
	"google.golang.org/grpc"                                  // gRPC сервер
	"google.golang.org/protobuf/proto"                        // сериализация protobuf-сообщений
)

// server реализует gRPC-сервис OrderService и хранит Kafka writer
type server struct {
	pb.UnimplementedOrderServiceServer
	kafkaWriter *kafka.Writer // writer для публикации сообщений в Kafka
}

// CreateOrder обрабатывает запрос на создание нового заказа
func (s *server) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
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

func main() {
	// Загружаем конфигурацию приложения (например, адрес gRPC, Kafka brokers и topic)
	appCfg := config.LoadConfig()

	// Разделяем строку с брокерами на слайс
	brokers := strings.Split(appCfg.KafkaBrokers, ",")

	// Создаём Kafka writer с конфигурацией брокеров и топика
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   appCfg.KafkaTopic,
	})
	defer writer.Close() // закрываем writer при завершении main

	// Создаём TCP listener для gRPC сервера
	lis, err := net.Listen("tcp", appCfg.OrderServiceAddr)
	if err != nil {
		log.Fatalf("listen: %v", err) // если порт занят или ошибка сети → завершаем приложение
	}

	// Создаём gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис OrderService
	pb.RegisterOrderServiceServer(s, &server{kafkaWriter: writer})

	log.Printf("Сервис заказов слушает на порту %s", appCfg.OrderServiceAddr)

	// Запускаем gRPC сервер и обрабатываем входящие запросы
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err) // если сервер упал → логируем и завершаем
	}
}
