package server

import (
	"context"
	"log"
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

type WorkerServer struct {
	reader    *kafka.Reader
	writer    *kafka.Writer
	dlqWriter *kafka.Writer
	rdb       *redis.Client
	ctx       context.Context
}

func NewWorkerServer(brokers []string, topic, dlqTopic, groupID, redisAddr string) *WorkerServer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})

	dlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   dlqTopic,
	})

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	return &WorkerServer{
		reader:    reader,
		writer:    writer,
		dlqWriter: dlqWriter,
		rdb:       rdb,
		ctx:       context.Background(),
	}
}

func (w *WorkerServer) Run() {
	defer w.reader.Close()
	defer w.writer.Close()
	defer w.dlqWriter.Close()

	for {
		msg, err := w.reader.FetchMessage(w.ctx)
		if err != nil {
			log.Printf("fetch error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var order pb.OrderRequest
		if err := proto.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("invalid message -> DLQ: %v", err)
			w.dlqWriter.WriteMessages(w.ctx, kafka.Message{Value: msg.Value})
			w.reader.CommitMessages(w.ctx, msg)
			continue
		}

		log.Printf("processing order %s", order.Id)
		time.Sleep(300 * time.Millisecond) // имитация обработки

		if strings.HasPrefix(order.Item, "fail") {
			retries := getRetries(msg)
			if retries < maxRetries {
				newMsg := kafka.Message{
					Key:     msg.Key,
					Value:   msg.Value,
					Headers: updateRetriesHeader(msg, retries+1),
				}
				if err := w.writer.WriteMessages(w.ctx, newMsg); err != nil {
					log.Printf("requeue failed: %v", err)
				} else {
					log.Printf("requeued %s (retry %d)", order.Id, retries+1)
				}
			} else {
				w.dlqWriter.WriteMessages(w.ctx, kafka.Message{Value: msg.Value})
				log.Printf("sent to DLQ: %s", order.Id)
			}
			w.reader.CommitMessages(w.ctx, msg)
			continue
		}

		res := &pb.ResultResponse{
			Item:   order.Item,
			Price:  order.Price,
			Status: "done",
		}
		b, _ := protojson.Marshal(res)
		if err := w.rdb.Set(w.ctx, "order:"+order.Id, b, 0).Err(); err != nil {
			log.Printf("redis set error: %v", err)
		}

		w.reader.CommitMessages(w.ctx, msg)
	}
}

// getRetries возвращает количество повторных попыток обработки сообщения из заголовка Kafka
func getRetries(msg kafka.Message) int {
	for _, h := range msg.Headers {
		if strings.ToLower(h.Key) == "retries" {
			v, _ := strconv.Atoi(string(h.Value))
			return v
		}
	}
	return 0
}

// updateRetriesHeader обновляет или добавляет заголовок retries с новым значением
func updateRetriesHeader(msg kafka.Message, retries int) []kafka.Header {
	headers := []kafka.Header{}
	found := false
	for _, h := range msg.Headers {
		if strings.ToLower(h.Key) == "retries" {
			headers = append(headers, kafka.Header{
				Key:   "retries",
				Value: []byte(strconv.Itoa(retries)),
			})
			found = true
		} else {
			headers = append(headers, h)
		}
	}
	if !found {
		headers = append(headers, kafka.Header{
			Key:   "retries",
			Value: []byte(strconv.Itoa(retries)),
		})
	}
	return headers
}