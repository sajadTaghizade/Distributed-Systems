package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MemoryEvent struct {
	EventType   string `json:"event_type"`
	Service     string `json:"service"`
	MemoryMB    uint64 `json:"memory_mb"`
	ThresholdMB uint64 `json:"threshold_mb"`
	Timestamp   string `json:"timestamp"`
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event MemoryEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Println("\n==================================================")
	fmt.Printf("URGENT ALERT %s\n", event.EventType)
	fmt.Println("==================================================")
	fmt.Printf("Service:      %s\n", event.Service)
	fmt.Printf("Memory Usage: %d MB\n", event.MemoryMB)
	fmt.Printf("Threshold:    %d MB\n", event.ThresholdMB)
	fmt.Printf("Timestamp:    %s\n", event.Timestamp)
	fmt.Println("==================================================\n")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event received by Subscriber"))
}

func main() {
	http.HandleFunc("/publish", handlePublish)

	log.Println("Pub/Sub Broker & Subscriber is listening on port 8083...")
	log.Fatal(http.ListenAndServe(":8083", nil))
}