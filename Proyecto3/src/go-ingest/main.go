package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pb "mumnk8s/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WarReport struct {
	Country          string `json:"country"`
	WarplanesInAir   int32  `json:"warplanes_in_air"`
	WarshipsInWater  int32  `json:"warships_in_water"`
	Timestamp        string `json:"timestamp"`
}

var grpcClient pb.WarReportServiceClient

func handleIngest(w http.ResponseWriter, r *http.Request) {
	var report WarReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	countryMap := map[string]pb.Countries{
		"usa": pb.Countries_usa,
		"rus": pb.Countries_rus,
		"chn": pb.Countries_chn,
		"esp": pb.Countries_esp,
		"gtm": pb.Countries_gtm,
	}

	pbCountry, ok := countryMap[strings.ToLower(report.Country)]
	if !ok {
		pbCountry = pb.Countries_countries_unknown
	}

	req := &pb.WarReportRequest{
		Country:         pbCountry,
		WarplanesInAir:  report.WarplanesInAir,
		WarshipsInWater: report.WarshipsInWater,
		Timestamp:       report.Timestamp,
	}

	res, err := grpcClient.SendReport(ctx, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("gRPC error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": res.Status})
}

func main() {
	grpcAddr := os.Getenv("GRPC_WRITER_ADDR")
	if grpcAddr == "" {
		grpcAddr = "go-writer.mumnk8s.svc.cluster.local:50051"
	}

	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpcClient = pb.NewWarReportServiceClient(conn)

	http.HandleFunc("/ingest", handleIngest)
	log.Println("Go Ingest API listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
