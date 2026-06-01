package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type LoginArgs struct {
	Username string
	Password string
}

type LoginReply struct {
	Success bool
	Message string
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthService struct {
	Users []User
}

func hashPassword(plainText string) string {
	hash := sha256.New()
	hash.Write([]byte(plainText))
	return hex.EncodeToString(hash.Sum(nil))
}

func (a *AuthService) Login(args *LoginArgs, reply *LoginReply) error {
	incomingHashedPassword := hashPassword(args.Password)

	reply.Success = false
	reply.Message = "Invalid Username or Password!"

	for _, u := range a.Users {
		if u.Username == args.Username && u.Password == incomingHashedPassword {
			reply.Success = true
			reply.Message = "Login Successful!"
			break
		}
	}

	return nil
}

func main() {
	file, err := os.ReadFile("users.json")
	if err != nil {
		log.Fatal("error reading database:", err)
	}

	var loadedUsers []User
	err = json.Unmarshal(file, &loadedUsers)
	if err != nil {
		log.Fatal("error parsing database:", err)
	}

	auth := &AuthService{Users: loadedUsers}
	rpc.Register(auth)

	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal("Error starting RPC server:", err)
	}

	fmt.Println("VM2 (Auth RPC Server) is running on port 8082...")
	fmt.Println("Security Note: Using SHA-256 password hashing. Users loaded into memory.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}