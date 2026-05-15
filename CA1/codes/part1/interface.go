package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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

	reqFile, err := os.OpenFile(reqPipe, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		fmt.Println("Worker is not running.")
		return
	}
	defer reqFile.Close()

	resFile, err := os.OpenFile(resPipe, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		fmt.Println("Failed to open response pipe.")
		return
	}
	defer resFile.Close()

	encoder := json.NewEncoder(reqFile)
	decoder := json.NewDecoder(resFile)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		parts := strings.Fields(text)
		if len(parts) != 3 {
			fmt.Println("Invalid input format. Use: OP A B")
			continue
		}

		op := strings.ToUpper(parts[0])
		a, errA := strconv.ParseFloat(parts[1], 64)
		b, errB := strconv.ParseFloat(parts[2], 64)

		if errA != nil || errB != nil {
			fmt.Println("Invalid numeric arguments.")
			continue
		}

		req := Request{
			Operation: op,
			A:         a,
			B:         b,
		}

		err = encoder.Encode(req)
		if err != nil {
			fmt.Println("Failed to send request.")
			break
		}

		var res Response
		err = decoder.Decode(&res)
		if err != nil {
			fmt.Println("Failed to read response.")
			break
		}

		if res.Status == "OK" {
			fmt.Println("Result:", res.Result)
		} else {
			fmt.Println("Error:", res.Error)
		}
	}
}
