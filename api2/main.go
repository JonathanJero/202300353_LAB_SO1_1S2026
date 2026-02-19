package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CONFIGURACIÓN DE IPs Y PUERTOS
const URL_API1 = "http://localhost:8080/health"        // API1 en la misma VM1
const URL_API3 = "http://192.168.122.5:8080/health"    // API3 en la VM2, verificar esta IP

// Estructuras JSON requeridas
type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	VM        string `json:"VM"`
	Carnet    string `json:"carnet"`
}

type CallResponse struct {
	ApiName    string `json:"apiname"`
	Message    string `json:"message"`
	Connection bool   `json:"connection"`
	Carnet     string `json:"carnet"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "UP",
		Message:   "API2 is Ready",
		Timestamp: time.Now().Format(time.RFC3339),
		VM:        "VM1",
		Carnet:    "202300353",
	})
}

func realizarLlamada(w http.ResponseWriter, url string, targetApiName string, targetVM string) {
	w.Header().Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)

	funciona := false
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var h HealthResponse
		json.Unmarshal(body, &h)
		if resp.StatusCode == 200 && h.Status == "UP" {
			funciona = true
		}
	}

	msg := fmt.Sprintf("ERROR: The %s located on the %s is not working", targetApiName, targetVM)
	if funciona {
		msg = fmt.Sprintf("The %s located on the %s is working", targetApiName, targetVM)
	}

	json.NewEncoder(w).Encode(CallResponse{
		ApiName:    targetApiName,
		Message:    msg,
		Connection: funciona,
		Carnet:     "202300353",
	})
}

// Endpoints requeridos
func callApi1Handler(w http.ResponseWriter, r *http.Request) {
	realizarLlamada(w, URL_API1, "API1", "VM1")
}

func callApi3Handler(w http.ResponseWriter, r *http.Request) {
	realizarLlamada(w, URL_API3, "API3", "VM2")
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api2/202300353/call-api1", callApi1Handler)
	http.HandleFunc("/api2/202300353/call-api3", callApi3Handler)

	fmt.Println("API2 corriendo en puerto 8081 (VM1)...")
	http.ListenAndServe(":8081", nil)
}