# VM1: Web Server

This virtual machine acts as the client-facing web server. It provides the HTML login interface, communicates with the Auth VM (VM2) via JSON-RPC, fetches assets from the File VM (VM3), and publishes memory alerts to the Pub/Sub Broker.

## Configuration
This service reads the following environment variables (with localhost fallbacks):
- `AUTH_VM_IP` (Default: 127.0.0.1:8082)
- `FILE_VM_IP` (Default: 127.0.0.1:8081)
- `PUBSUB_IP`  (Default: 127.0.0.1:8083)

## Setup & Run
1. Open a terminal in this directory.
2. Initialize the Go module (run once):
   ```bash
   go mod init webvm
3. the server (you can inject your own IPs here)
export AUTH_VM_IP="<AUTH_VM_IP>:8082"
export FILE_VM_IP="<FILE_VM_IP>:8081"
export PUBSUB_IP="<PUBSUB_VM_IP>:8083"
go run main.go