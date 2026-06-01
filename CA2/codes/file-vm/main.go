package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
)

type FileArgs struct {
	Category string 
	Filename string
}

type FileReply struct {
	Data []byte
}

type FileService struct{}

func (s *FileService) GetFile(args *FileArgs, reply *FileReply) error {
	safeFilename := filepath.Base(args.Filename)
	
	var basePath string
	if args.Category == "images" {
		basePath = "./images"
	} else {
		basePath = "./files"
	}

	filePath := filepath.Join(basePath, safeFilename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("file not found or unreadable: %v", err)
	}

	reply.Data = data
	return nil
}

func main() {
	fileService := new(FileService)
	rpc.Register(fileService)

	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Error starting File RPC server:", err)
	}
	
	fmt.Println("VM3 (File RPC Server) is running on port 8081...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}