//go:build integration

package test

import (
	"context"
	"log"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-portfolio/order-pipeline/internal/config"
	"github.com/go-portfolio/order-pipeline/internal/order"
	pb "github.com/go-portfolio/order-pipeline/proto"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestCreateOrder_Integration(t *testing.T) {
	appCfg := loadConfig()

	// Поднимаем gRPC сервер
	lis, err := net.Listen("tcp", appCfg.OrderServiceAddr)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	brokers := strings.Split(appCfg.KafkaBrokers, ",")
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    appCfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &order.Server{KafkaWriter: writer})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()
	defer s.Stop()

	// Даём серверу подняться
	time.Sleep(500 * time.Millisecond)

	// gRPC клиент
	conn, err := grpc.Dial(appCfg.OrderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	// Тестируем метод CreateOrder
	resp, err := client.CreateOrder(context.Background(), &pb.OrderRequest{
		Id:   "123",
		Item: "abc",
		Price:  2,
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if resp.GetStatus() != "accepted" {
		t.Errorf("unexpected response status: %s", resp.GetStatus())
	}

	// Тут можно добавить Kafka reader и прочитать сообщение, если нужно
}

// loadConfig загружает конфигурацию из файла, определяя путь до текущего файла
func loadConfig() *config.Config {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("не удалось определить путь до конфигурации")
	}
	return config.Load(filename)
}

