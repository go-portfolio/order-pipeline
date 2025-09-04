package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/stretchr/testify/require" // для удобных утверждений в тестах
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // используется для gRPC без TLS
)

// Константы для адресов сервисов и тестового заказа
const (
	orderServiceAddr = "orderreceiver:50051" // адрес gRPC сервиса OrderService
	cacheServiceAddr = "ordercache:50052"   // адрес gRPC сервиса CacheService
	testOrderID      = "e2e-test-1"         // уникальный ID тестового заказа
)

func TestFullPipeline(t *testing.T) {
	// Создаем контекст с таймаутом, чтобы тест не висел вечно
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() // отменяем контекст в конце функции

	// -----------------------------
	// 1️⃣ Подключение к OrderService и создание заказа
	// -----------------------------
	connOrder, err := grpc.Dial(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)          // проверяем, что соединение прошло успешно
	defer connOrder.Close()           // закрываем соединение после теста

	orderClient := pb.NewOrderServiceClient(connOrder) // создаем gRPC клиента для OrderService
	orderReq := &pb.OrderRequest{
		Id:    testOrderID, // уникальный идентификатор заказа
		Item:  "book",      // товар
		Price: 42,          // цена
	}

	respOrder, err := orderClient.CreateOrder(ctx, orderReq) // отправляем заказ
	require.NoError(t, err)                                  // проверяем, что ошибка отсутствует
	require.Equal(t, "accepted", respOrder.Status)           // ожидаем, что статус заказа accepted
	t.Logf("Order sent: %v", orderReq)                       // логируем отправленный заказ

	// -----------------------------
	// 2️⃣ Ждем обработки worker и получения результата из CacheService
	// -----------------------------
	var cacheResp *pb.ResultResponse
	retries := 5 // количество попыток опроса CacheService
	for i := 0; i < retries; i++ {
		// Подключаемся к CacheService
		connCache, err := grpc.Dial(cacheServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		cacheClient := pb.NewCacheServiceClient(connCache)

		// Запрашиваем результат заказа по ID
		cacheResp, err = cacheClient.GetOrderResult(ctx, &pb.ResultRequest{Id: testOrderID})
		connCache.Close() // закрываем соединение после каждого запроса

		if err == nil {
			break // если результат получен, выходим из цикла
		}
		time.Sleep(1 * time.Second) // ждем перед следующей попыткой
	}

	// Проверяем, что CacheService вернул результат
	require.NoError(t, err, "cache service did not return result in time")

	// Проверяем, что данные из CacheService соответствуют отправленному заказу
	require.Equal(t, orderReq.Item, cacheResp.Item)
	require.Equal(t, orderReq.Price, cacheResp.Price)
	require.Equal(t, "done", cacheResp.Status)
	t.Logf("Order processed successfully: %v", cacheResp) // логируем успешную обработку
}
