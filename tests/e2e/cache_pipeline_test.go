package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	orderServiceAddr = "orderservice:50051"
	cacheServiceAddr = "cacheservice:50052"
	testOrderID      = "e2e-test-1"
)

func TestFullPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1️⃣ Подключение к OrderService
	connOrder, err := grpc.NewClient(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// 2️⃣ Ждем обработки worker с polling
	var cacheResp *pb.ResultResponse
	retries := 5
	for i := 0; i < retries; i++ {
		connCache, err := grpc.NewClient(cacheServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		cacheClient := pb.NewCacheServiceClient(connCache)

		cacheResp, err = cacheClient.GetOrderResult(ctx, &pb.ResultRequest{Id: testOrderID})
		connCache.Close()

		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	require.NoError(t, err, "cache service did not return result in time")
	require.Equal(t, orderReq.Item, cacheResp.Item)
	require.Equal(t, orderReq.Price, cacheResp.Price)
	require.Equal(t, "done", cacheResp.Status)
	t.Logf("Order processed successfully: %v", cacheResp)
}
