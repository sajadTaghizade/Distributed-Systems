package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/rpc/jsonrpc"
	"os"
	"runtime"
	"strconv"
	"time"
)

type LoginArgs struct {
	Username string
	Password string
}

type LoginReply struct {
	Success bool
	Message string
}

type FileArgs struct {
	Category string
	Filename string
}

type FileReply struct {
	Data []byte
}

type MemoryEvent struct {
	EventType   string `json:"event_type"`
	Service     string `json:"service"`
	MemoryMB    uint64 `json:"memory_mb"`
	ThresholdMB uint64 `json:"threshold_mb"`
	Timestamp   string `json:"timestamp"`
}

var (
	authVMAddress string
	fileVMAddress string
	pubSubAddress string
	templates     = template.Must(template.ParseGlob("templates/*.html"))
	
	memoryLeakSlice [][]byte
)

func main() {
	authVMAddress = os.Getenv("AUTH_VM_IP")
	fileVMAddress = os.Getenv("FILE_VM_IP")
	pubSubAddress = os.Getenv("PUBSUB_IP")

	if authVMAddress == "" { authVMAddress = "127.0.0.1:8082" }
	if fileVMAddress == "" { fileVMAddress = "127.0.0.1:8081" }
	if pubSubAddress == "" { pubSubAddress = "127.0.0.1:8083" }

	go startMemoryMonitor()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/view-image", viewImageHandler)
	http.HandleFunc("/consume-memory", consumeMemoryHandler) 
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("VM1 (Web Server) is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startMemoryMonitor() {
	ticker := time.NewTicker(3 * time.Second) 
	var thresholdMB uint64 = 300              

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		currentMemMB := m.Alloc / 1024 / 1024

		if currentMemMB > thresholdMB {
			publishMemoryEvent(currentMemMB, thresholdMB)
		}
	}
}

func publishMemoryEvent(currentMem, threshold uint64) {
	event := MemoryEvent{
		EventType:   "HIGH_MEMORY_USAGE",
		Service:     "web-server",
		MemoryMB:    currentMem,
		ThresholdMB: threshold,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Println("Error marshaling memory event:", err)
		return
	}

	url := fmt.Sprintf("http://%s/publish", pubSubAddress)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to connect to Subscriber at %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[Pub] Alert sent! Current Memory: %d MB\n", currentMem)
}

func consumeMemoryHandler(w http.ResponseWriter, r *http.Request) {
	mbStr := r.URL.Query().Get("mb")
	mb := 50 
	if mbStr != "" {
		if val, err := strconv.Atoi(mbStr); err == nil {
			mb = val
		}
	}

	dummy := make([]byte, mb*1024*1024)
	for i := range dummy {
		dummy[i] = 1
	}
	
	memoryLeakSlice = append(memoryLeakSlice, dummy)

	fmt.Fprintf(w, "Allocated %d MB of memory successfully. Check Subscriber terminal!", mb)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	client, err := jsonrpc.Dial("tcp", authVMAddress)
	if err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Auth Service Unavailable"})
		return
	}
	defer client.Close()

	args := LoginArgs{Username: username, Password: password}
	var reply LoginReply

	err = client.Call("AuthService.Login", &args, &reply)
	if err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "RPC Communication Error"})
		return
	}

	if reply.Success {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": reply.Message})
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "dashboard.html", nil)
}

func viewImageHandler(w http.ResponseWriter, r *http.Request) {
	client, err := jsonrpc.Dial("tcp", fileVMAddress)
	if err != nil {
		http.Error(w, "File Service Unavailable", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	args := FileArgs{Category: "images", Filename: "secret.jpg"}
	var reply FileReply

	err = client.Call("FileService.GetFile", &args, &reply)
	if err != nil {
		http.Error(w, "Error retrieving file via RPC", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(reply.Data)
}