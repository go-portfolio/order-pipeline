package main

import (
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"
	pb "github.com/go-portfolio/order-pipeline/proto"
)

func main() {
	// Создаём структуру OrderRequest
	order := &pb.OrderRequest{
		Id:    "123",
		Item:  "Laptop",
		Price: 1500,
	}

	// Сериализуем в protobuf-байты
	data, err := proto.Marshal(order)
	if err != nil {
		log.Fatalf("Ошибка сериализации: %v", err)
	}

	fmt.Printf("Сериализованные байты: %x\n", data)

	// Десериализуем обратно
	var parsed pb.OrderRequest
	if err := proto.Unmarshal(data, &parsed); err != nil {
		log.Fatalf("Ошибка десериализации: %v", err)
	}

	// Проверка вывода
	fmt.Printf("Распарсено: ID=%s, Item=%s, Price=%d\n", parsed.Id, parsed.Item, parsed.Price)
}
