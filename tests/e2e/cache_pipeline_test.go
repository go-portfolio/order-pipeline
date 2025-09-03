package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Параметры подключения к сервисам
const (
	orderServiceAddr = "localhost:50051"
	cacheServiceAddr = "localhost:50052"
	testOrderID      = "e2e-test-1"
)

// TestFullPipeline проверяет, что заказ проходит полный путь:
// orderservice -> Kafka -> worker -> Redis -> cacheservice
func TestFullPipeline(t *testing.T) {
	ctx := context.Background()

	// 1️⃣ Отправляем заказ в OrderService
	connOrder, err := grpc.Dial(orderServiceAddr, grpc.WithInsecure())
	require.NoError(t, err)
	defer connOrder.Close()

	orderClient := pb.NewOrderServiceClient(connOrder)
	orderReq := &pb.OrderRequest{
		Id:    testOrderID,
		Item:  "book",
		Price: 42,
	}

	respOrder, err := orderClient.CreateOrder(ctx, orderReq)
	require.NoError(t, err)
	require.Equal(t, "accepted", respOrder.Status)

	t.Logf("Order sent: %v", orderReq)

	// 2️⃣ Ждем обработки worker (обычно достаточно 1-2 секунды)
	time.Sleep(2 * time.Second)

	// 3️⃣ Проверяем, что результат доступен в CacheService
	connCache, err := grpc.Dial(cacheServiceAddr, grpc.WithInsecure())
	require.NoError(t, err)
	defer connCache.Close()

	cacheClient := pb.NewCacheServiceClient(connCache)
	cacheResp, err := cacheClient.GetOrderResult(ctx, &pb.ResultRequest{Id: testOrderID})
	require.NoError(t, err)

	// 4️⃣ Проверяем, что данные совпадают
	require.Equal(t, orderReq.Item, cacheResp.Item)
	require.Equal(t, orderReq.Price, cacheResp.Price)
	require.Equal(t, "done", cacheResp.Status)

	t.Logf("Order processed successfully: %v", cacheResp)
}
