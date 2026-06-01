package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MemoryEvent struct {
	EventType   string `json:"event_type"`
	Service     string `json:"service"`
	MemoryMB    uint64 `json:"memory_mb"`
	ThresholdMB uint64 `json:"threshold_mb"`
	Timestamp   string `json:"timestamp"`
}

func main() {
	mockEvent := MemoryEvent{
		EventType:   "HIGH_MEMORY_USAGE",
		Service:     "standalone-publisher-test",
		MemoryMB:    345,
		ThresholdMB: 300,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	jsonData, err := json.Marshal(mockEvent)
	if err != nil {
		log.Fatal("Error marshaling event:", err)
	}

	url := "http://127.0.0.1:8083/publish"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to connect to Subscriber: %v\n", err)
	}
	defer resp.Body.Close()

	fmt.Println("Mock event successfully published to Subscriber!")
}