package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)


type DataItem struct {
	Value     string `json:"value"`
	Version   int    `json:"version"`
	UpdatedBy string `json:"updated_by"`
}

type ClientRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Consistency string `json:"consistency"` 
}


type ReplicateRequest struct {
	Key  string   `json:"key"`
	Data DataItem `json:"data"`
}


type Config struct {
	ID    string   `json:"id"`
	Port  string   `json:"port"`
	Peers []string `json:"peers"`
}

var (
	store = make(map[string]DataItem)
	mu    sync.RWMutex

	id      string
	port    string
	peers   []string
	delayMs int
)

func main() {
	flag.StringVar(&id, "id", "replica1", "Replica ID")
	flag.StringVar(&port, "port", "8081", "Port to listen on")
	peersStr := flag.String("peers", "", "Comma-separated peer URLs, e.g. http://localhost:8082,http://localhost:8083")
	flag.IntVar(&delayMs, "delay", 0, "Artificial delay (ms) applied before sending each replication message")
	configPath := flag.String("config", "", "Optional JSON config file; its values override -id/-port/-peers")
	flag.Parse()

	if *peersStr != "" {
		peers = strings.Split(*peersStr, ",")
	}


	if *configPath != "" {
		if err := loadConfig(*configPath); err != nil {
			log.Fatalf("failed to load config %q: %v", *configPath, err)
		}
	}

	http.HandleFunc("/put", handlePut)
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/internal/replicate", handleReplicate)

	log.Printf("starting %s on :%s | peers=%v | delay=%dms | quorum=%d", id, port, peers, delayMs, quorum())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadConfig(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return err
	}
	if c.ID != "" {
		id = c.ID
	}
	if c.Port != "" {
		port = c.Port
	}
	if len(c.Peers) > 0 {
		peers = c.Peers
	}
	return nil
}


func quorum() int {
	return (len(peers)+1)/2 + 1
}


func handlePut(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	mu.Lock()
	current := store[req.Key]
	newData := DataItem{
		Value:     req.Value,
		Version:   current.Version + 1,
		UpdatedBy: id,
	}
	store[req.Key] = newData
	mu.Unlock()

	replReq := ReplicateRequest{Key: req.Key, Data: newData}

	if req.Consistency == "strong" {
		acks := 1 
		var wg sync.WaitGroup
		var ackMu sync.Mutex
		for _, peer := range peers {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				if sendReplication(p, replReq) {
					ackMu.Lock()
					acks++
					ackMu.Unlock()
				}
			}(peer)
		}
		wg.Wait()

		q := quorum()
		total := len(peers) + 1
		if acks >= q {
			log.Printf("[PUT] key=%s value=%q -> v%d (strong) acks=%d/%d quorum=%d OK",
				req.Key, req.Value, newData.Version, acks, total, q)
			writeJSON(w, http.StatusOK, map[string]any{
				"status":           "success",
				"consistency":      "strong",
				"key":              req.Key,
				"version":          newData.Version,
				"replicas_updated": acks,
				"quorum":           q,
			})
		} else {
			log.Printf("[PUT] key=%s value=%q -> v%d (strong) acks=%d/%d quorum=%d FAILED",
				req.Key, req.Value, newData.Version, acks, total, q)
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"status":           "error",
				"message":          "failed to achieve quorum",
				"consistency":      "strong",
				"key":              req.Key,
				"version":          newData.Version,
				"replicas_updated": acks,
				"quorum":           q,
			})
		}
		return
	}

	for _, peer := range peers {
		go sendReplication(peer, replReq)
	}
	log.Printf("[PUT] key=%s value=%q -> v%d (eventual) replicating async to %d peer(s)",
		req.Key, req.Value, newData.Version, len(peers))
	writeJSON(w, http.StatusOK, map[string]any{
		"status":           "success",
		"consistency":      "eventual",
		"key":              req.Key,
		"version":          newData.Version,
		"replicas_updated": 1,
		"note":             "asynchronous replication in progress",
	})
}


func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	consistency := r.URL.Query().Get("consistency")

	mu.RLock()
	localData, exists := store[key]
	mu.RUnlock()

	if consistency != "strong" {
		if !exists {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "key not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"key": key, "value": localData.Value, "version": localData.Version, "served_by": id,
		})
		return
	}

	highestData := localData
	acks := 1 
	

	if acks >= quorum() {
        if highestData.Version == 0 {
            writeJSON(w, http.StatusNotFound, map[string]any{"error": "key not found"})
            return
        }
		writeJSON(w, http.StatusOK, map[string]any{
			"key": key, "value": highestData.Value, "version": highestData.Version, "served_by": id, "note": "strong read",
		})
	} else {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to achieve read quorum"})
	}
}

func handleReplicate(w http.ResponseWriter, r *http.Request) {
	var req ReplicateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	current, exists := store[req.Key]
	applied := !exists ||
		req.Data.Version > current.Version ||
		(req.Data.Version == current.Version && req.Data.UpdatedBy > current.UpdatedBy)
	if applied {
		store[req.Key] = req.Data
	}
	mu.Unlock()

	if applied {
		log.Printf("[REPLICATE] recv key=%s v%d from %s -> APPLIED", req.Key, req.Data.Version, req.Data.UpdatedBy)
	} else {
		log.Printf("[REPLICATE] recv key=%s v%d from %s -> rejected (kept v%d by %s)",
			req.Key, req.Data.Version, req.Data.UpdatedBy, current.Version, current.UpdatedBy)
	}
	writeJSON(w, http.StatusOK, map[string]any{"applied": applied, "key": req.Key})
}


func sendReplication(peer string, req ReplicateRequest) bool {
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	body, _ := json.Marshal(req)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(peer+"/internal/replicate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[REPLICATE] send key=%s v%d to %s -> ERROR: %v", req.Key, req.Data.Version, peer, err)
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == http.StatusOK
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
