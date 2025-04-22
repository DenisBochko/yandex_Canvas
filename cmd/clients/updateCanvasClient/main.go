package main

import (
	"context"
	"fmt"
	"log"
	"time"

	canavasv1 "gitlab.crja72.ru/golang/2025/spring/course/projects/go6/contracts/gen/go/canvas"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := canavasv1.NewCanvasClient(conn)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		"uid":      "e3ff6d93-899c-4150-860c-e3ed5e361563",
		"verified": "true",
	}))
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Запрос
	req := &canavasv1.UpdateCanvasRequest{
		CanvasId: "d16c8b70-e3ef-4716-b8c9-65d9ecfdcc82",
		Name:     "New-test1",
		Privacy:  "public",
	}

	res, err := client.UpdateCanvas(ctx, req)
	if err != nil {
		log.Fatalf("ошибка при вызове UpdateCanvas: %v", err)
	}

	fmt.Printf("Успешно вызван UpdateCanvas, canvasID: %s\n", res.GetCanvasId())
}
