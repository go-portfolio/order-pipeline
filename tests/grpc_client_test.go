package tests

import (
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestOrderServiceConnection(t *testing.T) {
	addr := "127.0.0.1:50051"

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()
}
