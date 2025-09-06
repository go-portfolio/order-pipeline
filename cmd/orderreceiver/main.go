package main

import (
	"log"
	"net"
	"strings"

	"github.com/go-portfolio/order-pipeline/internal/config" // пакет для загрузки конфигурации приложения
	"github.com/go-portfolio/order-pipeline/internal/server"
	pb "github.com/go-portfolio/order-pipeline/proto" // сгенерированные protobuf файлы для OrderService
	"github.com/segmentio/kafka-go"                   // клиент Kafka для записи сообщений
	"google.golang.org/grpc"                          // gRPC сервер
	"google.golang.org/grpc/reflection"
	// сериализация protobuf-сообщений
)

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
	pb.RegisterOrderServiceServer(s, server.NewOrderServer(writer))

	log.Printf("Сервис заказов слушает на порту %s", appCfg.OrderServiceAddr)

	// Включаем reflection
	reflection.Register(s)

	// Запускаем gRPC сервер и обрабатываем входящие запросы
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err) // если сервер упал → логируем и завершаем
	}
}
