package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxRetries = 3

func main() {
	brokersEnv := getenv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")
	topic := getenv("KAFKA_TOPIC", "orders")
	dlqTopic := getenv("DLQ_TOPIC", "orders-dlq")
	groupID := getenv("WORKER_GROUP", "worker-group")
	redisAddr := getenv("REDIS_ADDR", "redis:6379")

	reader := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, GroupID: groupID, Topic: topic})
	defer reader.Close()

	writer := kafka.NewWriter(kafka.WriterConfig{Brokers: brokers, Topic: topic})
	defer writer.Close()

	dlqWriter := kafka.NewWriter(kafka.WriterConfig{Brokers: brokers, Topic: dlqTopic})
	defer dlqWriter.Close()

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("fetch error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var order pb.OrderRequest
		if err := proto.Unmarshal(msg.Value, &order); err != nil {
			// can't parse -> DLQ
			log.Printf("invalid message -> DLQ: %v", err)
			dlqWriter.WriteMessages(ctx, kafka.Message{Value: msg.Value})
			reader.CommitMessages(ctx, msg)
			continue
		}

		log.Printf("processing order %s", order.Id)
		time.Sleep(300 * time.Millisecond)

		// Демонстрация ошибки: если item начинается с "fail" — симулируем сбой
		if strings.HasPrefix(order.Item, "fail") {
			retries := getRetries(msg)
			if retries < maxRetries {
				newMsg := kafka.Message{Key: msg.Key, Value: msg.Value, Headers: updateRetriesHeader(msg, retries+1)}
				if err := writer.WriteMessages(ctx, newMsg); err != nil {
					log.Printf("requeue failed: %v", err)
				} else {
					log.Printf("requeued %s (retry %d)", order.Id, retries+1)
				}
			} else {
				dlqWriter.WriteMessages(ctx, kafka.Message{Value: msg.Value})
				log.Printf("sent to DLQ: %s", order.Id)
			}
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Успешная обработка -> пишем результат в Redis (protojson)
		res := &pb.ResultResponse{Item: order.Item, Price: order.Price, Status: "done"}
		b, _ := protojson.Marshal(res)
		if err := rdb.Set(ctx, "order:"+order.Id, b, 0).Err(); err != nil {
			log.Printf("redis set error: %v", err)
		}
		reader.CommitMessages(ctx, msg)
	}
}

func getRetries(msg kafka.Message) int {
	for _, h := range msg.Headers {
		if strings.ToLower(h.Key) == "retries" {
			v, _ := strconv.Atoi(string(h.Value))
			return v
		}
	}
	return 0
}

func updateRetriesHeader(msg kafka.Message, retries int) []kafka.Header {
	headers := []kafka.Header{}
	found := false
	for _, h := range msg.Headers {
		if strings.ToLower(h.Key) == "retries" {
			headers = append(headers, kafka.Header{Key: "retries", Value: []byte(strconv.Itoa(retries))})
			found = true
		} else {
			headers = append(headers, h)
		}
	}
	if !found {
		headers = append(headers, kafka.Header{Key: "retries", Value: []byte(strconv.Itoa(retries))})
	}
	return headers
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
