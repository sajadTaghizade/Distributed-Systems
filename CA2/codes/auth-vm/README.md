# VM2: Authentication Server

This virtual machine handles user authentication via JSON-RPC. It securely reads user credentials from the local `users.json` database into memory at startup, and validates login requests sent by the Web VM. 

## Features
- **RPC Service:** Implements standard JSON-RPC over TCP.
- **Security (Bonus):** Passwords are hashed using SHA-256.
- **In-Memory Database:** Parses `users.json` once at startup for optimized lookup times.
- **Data Isolation:** Ensures the Web VM (VM1) has no direct access to user files.

## Setup & Run

1. Open a terminal in this directory.
2. Initialize the Go module (run once):
   ```bash
   go mod init authvm