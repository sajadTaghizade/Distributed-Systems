package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type ComputeResult struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
	Result    float64 `json:"result,omitempty"`
}

type ErrorResult struct {
	Error string `json:"error"`
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("Method: %s\tPath: %s\tQuery: %s\tDuration: %v", r.Method, r.URL.Path, r.URL.RawQuery, time.Since(start))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResult{Error: "Method Not Allowed"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func computeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResult{Error: "Method Not Allowed"})
		return
	}

	op := r.URL.Query().Get("op")
	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	if op == "" || aStr == "" || bStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResult{Error: "Missing parameters"})
		return
	}

	a, errA := strconv.ParseFloat(aStr, 64)
	b, errB := strconv.ParseFloat(bStr, 64)
	if errA != nil || errB != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResult{Error: "Non-numeric values for a or b"})
		return
	}

	var result float64
	switch op {
	case "add":
		result = a + b
	case "sub":
		result = a - b
	case "mul":
		result = a * b
	case "div":
		if b == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResult{Error: "Division by zero"})
			return
		}
		result = a / b
	case "mod":
		if b == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResult{Error: "Modulo by zero"})
			return
		}
		result = math.Mod(a, b)
	case "pow":
		result = math.Pow(a, b)
	case "max":
		result = math.Max(a, b)
	case "min":
		result = math.Min(a, b)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResult{Error: "Invalid op"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ComputeResult{
		Operation: op,
		A:         a,
		B:         b,
		Result:    result,
	})
}

func main() {
	http.HandleFunc("/health", loggingMiddleware(healthHandler))
	http.HandleFunc("/compute", loggingMiddleware(computeHandler))
	log.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
