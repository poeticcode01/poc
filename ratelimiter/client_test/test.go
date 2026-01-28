package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func main() {
	const (
		nRequests    = 100
		requestDelay = 10 * time.Millisecond // Delay between requests
	)

	fmt.Printf("Sending %d requests to http://localhost:8085/rate-limit\n", nRequests)

	var wg sync.WaitGroup
	for i := 0; i < nRequests; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()
			resp, err := http.Get("http://localhost:8085/rate-limit")
			if err != nil {
				fmt.Printf("Request %d failed: %v\n", reqNum, err)
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response body for request %d: %v\n", reqNum, err)
				return
			}
			bodyString := string(bodyBytes)
			fmt.Printf("Request %d: Status %s, Body: %s\n", reqNum, resp.Status, bodyString)
		}(i + 1)

		time.Sleep(requestDelay)
	}

	wg.Wait()
	fmt.Println("All requests sent.")
}
