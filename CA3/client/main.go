package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)



func main() {
	action := flag.String("action", "", "Operation: get or put")
	node := flag.String("node", "http://localhost:8081", "Replica URL to contact")
	key := flag.String("key", "", "Key to get/put")
	value := flag.String("value", "", "Value to put")
	consistency := flag.String("consistency", "eventual", "Consistency model for put: eventual or strong")
	flag.Parse()

	switch *action {
	case "put":
		if *key == "" || *value == "" {
			fmt.Println("error: -key and -value are required for put")
			os.Exit(1)
		}
		body, _ := json.Marshal(map[string]string{
			"key":         *key,
			"value":       *value,
			"consistency": *consistency,
		})
		start := time.Now()
		resp, err := http.Post(*node+"/put", "application/json", bytes.NewBuffer(body))
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("PUT %s=%s [%s] -> %s FAILED after %s: %v\n",
				*key, *value, *consistency, *node, elapsed.Round(time.Millisecond), err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		out, _ := io.ReadAll(resp.Body)
		fmt.Printf("PUT %s=%s [%s] -> %s | %.1f ms | HTTP %d\n",
			*key, *value, *consistency, *node, ms(elapsed), resp.StatusCode)
		fmt.Printf("  response: %s\n", trim(out))

	case "get":
		if *key == "" {
			fmt.Println("error: -key is required for get")
			os.Exit(1)
		}
		start := time.Now()
		resp, err := http.Get(fmt.Sprintf("%s/get?key=%s", *node, *key))
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("GET %s -> %s FAILED after %s: %v\n",
				*key, *node, elapsed.Round(time.Millisecond), err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		out, _ := io.ReadAll(resp.Body)
		fmt.Printf("GET %s -> %s | %.1f ms | HTTP %d\n", *key, *node, ms(elapsed), resp.StatusCode)
		fmt.Printf("  response: %s\n", trim(out))

	default:
		fmt.Println("usage: client -action=get|put -node=URL -key=K [-value=V] [-consistency=eventual|strong]")
		os.Exit(1)
	}
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000.0 }

func trim(b []byte) string {
	s := string(b)
	if n := len(s); n > 0 && s[n-1] == '\n' {
		s = s[:n-1]
	}
	return s
}
