package e2e

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/go-portfolio/order-pipeline/internal/config"
	pb "github.com/go-portfolio/order-pipeline/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const testOrderID = "e2e-test-1"

// getLocalIPs возвращает список всех IP адресов контейнера
func getLocalIPs() ([]string, error) {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips, nil
}

func TestFullPipelineWithIP(t *testing.T) {
	// Загружаем конфигурацию приложения
	appCfg := config.LoadConfig()
	fmt.Println(appCfg.CacheServiceAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ips, err := getLocalIPs()
	require.NoError(t, err)
	require.NotEmpty(t, ips)

	// Берем первый доступный IP (можно выбрать нужный по условию)
	localIP := ips[0]
	t.Logf("Using local container IP: %s", localIP)

	cacheServiceAddr := "cacheservice:50052"

	// 1️⃣ Подключение к OrderService
	connOrder, err := grpc.Dial(cacheServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	// 2️⃣ Polling для проверки результата в CacheService
	var cacheResp *pb.ResultResponse
	retries := 10
	for i := 0; i < retries; i++ {
		connCache, err := grpc.Dial(cacheServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
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
