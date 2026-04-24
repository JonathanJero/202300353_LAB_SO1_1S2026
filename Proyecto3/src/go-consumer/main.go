package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq.mumnk8s.svc.cluster.local:5672/"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "valkey.mumnk8s.svc.cluster.local:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Valkey Database: %v", err)
	}
	log.Println("Connected to Valkey Storage Machine successfully.")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ broker: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open an AMQP channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"wartweets", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue configuration: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a subscriber consumer: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for d := range msgs {
			uniqueID := fmt.Sprintf("report:%d", time.Now().UnixNano())
			err := rdb.Set(ctx, uniqueID, string(d.Body), 0).Err()
			if err != nil {
				log.Printf("Critical: Failed persisting payload inside Valkey: %v", err)
			} else {
				log.Printf("Successfully saved message: %s -> %s", uniqueID, d.Body)
			}
		}
	}()

	log.Printf("Go Application Subscriber connected cleanly. Monitoring telemetry traffic...")
	<-sigChan
	log.Println("Shutting down the subscriber unit.")
}
