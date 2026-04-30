package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type WarReport struct {
	Country          string `json:"country"`
	WarplanesInAir   int32  `json:"warplanes_in_air"`
	WarshipsInWater  int32  `json:"warships_in_water"`
	Timestamp        string `json:"timestamp"`
}

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
			var report WarReport
			if err := json.Unmarshal(d.Body, &report); err != nil {
				log.Printf("Error unmarshaling json: %v", err)
				continue
			}

			uniqueID := fmt.Sprintf("report:%d", time.Now().UnixNano())
			err := rdb.Set(ctx, uniqueID, string(d.Body), 0).Err()
			if err != nil {
				log.Printf("Critical: Failed persisting payload inside Valkey: %v", err)
			} else {
				log.Printf("Successfully saved message: %s -> %s", uniqueID, d.Body)
			}

			// AGREGACIONES PARA GRAFANA
			
			// 1 y 2: Max/Min Aviones
			currMaxA, _ := rdb.Get(ctx, "max_aviones").Int()
			if currMaxA == 0 || int(report.WarplanesInAir) > currMaxA { rdb.Set(ctx, "max_aviones", report.WarplanesInAir, 0) }
			currMinA, errMinA := rdb.Get(ctx, "min_aviones").Int()
			if errMinA != nil || int(report.WarplanesInAir) < currMinA { rdb.Set(ctx, "min_aviones", report.WarplanesInAir, 0) }

			// 3 y 4: Max/Min Barcos
			currMaxB, _ := rdb.Get(ctx, "max_barcos").Int()
			if currMaxB == 0 || int(report.WarshipsInWater) > currMaxB { rdb.Set(ctx, "max_barcos", report.WarshipsInWater, 0) }
			currMinB, errMinB := rdb.Get(ctx, "min_barcos").Int()
			if errMinB != nil || int(report.WarshipsInWater) < currMinB { rdb.Set(ctx, "min_barcos", report.WarshipsInWater, 0) }

			// 5 y 6: Top Paises
			rdb.HIncrBy(ctx, "top_aviones_pais", report.Country, int64(report.WarplanesInAir))
			rdb.HIncrBy(ctx, "top_barcos_pais", report.Country, int64(report.WarshipsInWater))

			// 7 y 8: Moda (Frecuencias y calculo en tiempo real)
			newFreqA := rdb.HIncrBy(ctx, "freq_aviones", fmt.Sprintf("%d", report.WarplanesInAir), 1).Val()
			maxFreqA, _ := rdb.Get(ctx, "max_freq_aviones").Int64()
			if newFreqA > maxFreqA {
				rdb.Set(ctx, "max_freq_aviones", newFreqA, 0)
				rdb.Set(ctx, "moda_aviones", report.WarplanesInAir, 0)
			}

			newFreqB := rdb.HIncrBy(ctx, "freq_barcos", fmt.Sprintf("%d", report.WarshipsInWater), 1).Val()
			maxFreqB, _ := rdb.Get(ctx, "max_freq_barcos").Int64()
			if newFreqB > maxFreqB {
				rdb.Set(ctx, "max_freq_barcos", newFreqB, 0)
				rdb.Set(ctx, "moda_barcos", report.WarshipsInWater, 0)
			}

			// 9, 10, 11: Datos exclusivos de RUS
			if report.Country == "rus" {
				rdb.Incr(ctx, "total_rus_reportes")
				// Usar Redis Streams (XADD) para que Grafana lo lea como TimeSeries automáticamente
				rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: "ts_rus_stream",
					Values: map[string]interface{}{
						"aviones": report.WarplanesInAir,
						"barcos":  report.WarshipsInWater,
					},
				})
			}
		}
	}()

	log.Printf("Go Application Subscriber connected cleanly. Monitoring telemetry traffic...")
	<-sigChan
	log.Println("Shutting down the subscriber unit.")
}
