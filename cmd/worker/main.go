package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-portfolio/order-pipeline/internal/config" // пакет для загрузки конфигурации приложения
	pb "github.com/go-portfolio/order-pipeline/proto"        // сгенерированные protobuf файлы
	"github.com/redis/go-redis/v9"                           // клиент Redis
	"github.com/segmentio/kafka-go"                           // клиент Kafka для чтения и записи сообщений
	"google.golang.org/protobuf/encoding/protojson"          // кодирование/декодирование protobuf в JSON
	"google.golang.org/protobuf/proto"                       // сериализация/десериализация protobuf
)

const maxRetries = 3 // максимальное количество попыток повторной обработки сообщения

func main() {
	// Загружаем конфигурацию приложения (Kafka, Redis, топики, группы)
	appCfg := config.LoadConfig()

	brokersEnv := appCfg.KafkaBrokers
	brokers := strings.Split(brokersEnv, ",") // список брокеров Kafka
	topic := appCfg.KafkaTopic                 // топик для новых заказов
	dlqTopic := appCfg.DlqTopic               // топик для сообщений, которые не удалось обработать (Dead Letter Queue)
	groupID := appCfg.WorkerGroup              // consumer group Kafka
	redisAddr := appCfg.RedisAddr              // адрес Redis

	// Создаём Kafka reader для чтения сообщений из топика заказов
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})
	defer reader.Close()

	// Kafka writer для переотправки сообщений в основной топик (retry)
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})
	defer writer.Close()

	// Kafka writer для Dead Letter Queue
	dlqWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   dlqTopic,
	})
	defer dlqWriter.Close()

	// Redis клиент для записи результатов заказов
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()

	// Основной цикл обработки сообщений
	for {
		// Получаем следующее сообщение из Kafka
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("fetch error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Десериализуем protobuf сообщение
		var order pb.OrderRequest
		if err := proto.Unmarshal(msg.Value, &order); err != nil {
			// Сообщение невалидное → отправляем в DLQ
			log.Printf("invalid message -> DLQ: %v", err)
			dlqWriter.WriteMessages(ctx, kafka.Message{Value: msg.Value})
			reader.CommitMessages(ctx, msg) // подтверждаем, что сообщение обработано
			continue
		}

		log.Printf("processing order %s", order.Id)
		time.Sleep(300 * time.Millisecond) // имитация обработки

		// Симуляция ошибки: если item начинается с "fail", обрабатываем retry
		if strings.HasPrefix(order.Item, "fail") {
			retries := getRetries(msg) // читаем текущее количество попыток
			if retries < maxRetries {
				// повторная отправка в основной топик с увеличенным счетчиком retries
				newMsg := kafka.Message{
					Key:     msg.Key,
					Value:   msg.Value,
					Headers: updateRetriesHeader(msg, retries+1),
				}
				if err := writer.WriteMessages(ctx, newMsg); err != nil {
					log.Printf("requeue failed: %v", err)
				} else {
					log.Printf("requeued %s (retry %d)", order.Id, retries+1)
				}
			} else {
				// превышено максимальное количество попыток → DLQ
				dlqWriter.WriteMessages(ctx, kafka.Message{Value: msg.Value})
				log.Printf("sent to DLQ: %s", order.Id)
			}
			reader.CommitMessages(ctx, msg)
			continue
		}

		// Успешная обработка → сохраняем результат в Redis
		res := &pb.ResultResponse{
			Item:   order.Item,
			Price:  order.Price,
			Status: "done",
		}
		b, _ := protojson.Marshal(res)       // сериализация protobuf в JSON для хранения в Redis
		if err := rdb.Set(ctx, "order:"+order.Id, b, 0).Err(); err != nil {
			log.Printf("redis set error: %v", err)
		}

		// Подтверждаем обработку сообщения в Kafka
		reader.CommitMessages(ctx, msg)
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
