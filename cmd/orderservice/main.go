package main

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/go-portfolio/order-pipeline/internal/config"
	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type server struct {
	pb.UnimplementedOrderServiceServer
	kafkaWriter *kafka.Writer
}

func (s *server) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	msg := kafka.Message{Value: b}
	if err := s.kafkaWriter.WriteMessages(ctx, msg); err != nil {
		return nil, err
	}
	return &pb.OrderResponse{Status: "accepted"}, nil
}

func main() {
	// Загружаем конфигурацию приложения из файла
	appCfg := config.LoadConfig()
	brokers := strings.Split(appCfg.KafkaBrokers, ",")
	writer := kafka.NewWriter(kafka.WriterConfig{Brokers: brokers, Topic: appCfg.KafkaTopic})
	defer writer.Close()

	lis, err := net.Listen("tcp", appCfg.OrderServiceAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &server{kafkaWriter: writer})
	log.Printf("OrderService listening on %s", appCfg.OrderServiceAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
