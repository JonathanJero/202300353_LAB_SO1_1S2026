package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	pb "mumnk8s/proto"
)

type server struct {
	pb.UnimplementedWarReportServiceServer
	rabbitChannel *amqp.Channel
	queueName     string
}

func (s *server) SendReport(ctx context.Context, in *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	log.Printf("Received from Ingest:%v", in)

	body, err := json.Marshal(map[string]interface{}{
		"country":           in.Country.String(),
		"warplanes_in_air":  in.WarplanesInAir,
		"warships_in_water": in.WarshipsInWater,
		"timestamp":         in.Timestamp,
	})
	if err != nil {
		log.Printf("Error marshalling to JSON: %s", err)
		return nil, err
	}

	err = s.rabbitChannel.PublishWithContext(ctx,
		"",          // exchange
		s.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
		return nil, err
	}

	return &pb.WarReportResponse{Status: "Message queued via RabbitMQ"}, nil
}

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq.mumnk8s.svc.cluster.local:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
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
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWarReportServiceServer(s, &server{
		rabbitChannel: ch,
		queueName:     q.Name,
	})
	log.Printf("Go Writer gRPC Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
