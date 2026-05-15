package main

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"syscall"
)

type Request struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

type Response struct {
	Status string  `json:"status"`
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

func main() {
	reqPipe := "/tmp/calc_req"
	resPipe := "/tmp/calc_res"

	os.Remove(reqPipe)
	os.Remove(resPipe)

	err := syscall.Mkfifo(reqPipe, 0666)
	if err != nil {
		log.Fatalf("Failed to create req pipe: %v", err)
	}
	err = syscall.Mkfifo(resPipe, 0666)
	if err != nil {
		log.Fatalf("Failed to create res pipe: %v", err)
	}

	defer os.Remove(reqPipe)
	defer os.Remove(resPipe)

	for {
		reqFile, err := os.OpenFile(reqPipe, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Printf("Req pipe open error: %v", err)
			continue
		}

		resFile, err := os.OpenFile(resPipe, os.O_WRONLY, os.ModeNamedPipe)
		if err != nil {
			reqFile.Close()
			log.Printf("Res pipe open error: %v", err)
			continue
		}

		decoder := json.NewDecoder(reqFile)
		encoder := json.NewEncoder(resFile)

		for {
			var req Request
			err := decoder.Decode(&req)
			if err != nil {
				break
			}

			var res Response
			switch req.Operation {
			case "ADD":
				res = Response{Status: "OK", Result: req.A + req.B}
			case "SUB":
				res = Response{Status: "OK", Result: req.A - req.B}
			case "MUL":
				res = Response{Status: "OK", Result: req.A * req.B}
			case "DIV":
				if req.B == 0 {
					res = Response{Status: "ERR", Error: "division by zero"}
				} else {
					res = Response{Status: "OK", Result: req.A / req.B}
				}
			case "MOD":
				if req.B == 0 {
					res = Response{Status: "ERR", Error: "modulo by zero"}
				} else {
					res = Response{Status: "OK", Result: math.Mod(req.A, req.B)}
				}
			case "POW":
				res = Response{Status: "OK", Result: math.Pow(req.A, req.B)}
			case "MAX":
				res = Response{Status: "OK", Result: math.Max(req.A, req.B)}
			case "MIN":
				res = Response{Status: "OK", Result: math.Min(req.A, req.B)}
			default:
				res = Response{Status: "ERR", Error: "unknown operation"}
			}

			if res.Status == "OK" {
				log.Printf("Request: %v %v %v -> Success: %v", req.Operation, req.A, req.B, res.Result)
			} else {
				log.Printf("Request: %v %v %v -> Error: %v", req.Operation, req.A, req.B, res.Error)
			}

			err = encoder.Encode(res)
			if err != nil {
				break
			}
		}

		reqFile.Close()
		resFile.Close()
	}
}
