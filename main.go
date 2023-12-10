package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       string  `json:"time_ns"`
}

func sortSequential(input [][]int) [][]int {
	for i := range input {
		sort.Ints(input[i])
	}
	return input
}

func sortConcurrent(input [][]int) [][]int {
	var wg sync.WaitGroup
	ch := make(chan []int, len(input))

	for _, arr := range input {
		wg.Add(1)
		go func(a []int) {
			sort.Ints(a)
			ch <- a
			wg.Done()
		}(arr)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var result [][]int
	for sortedArr := range ch {
		result = append(result, sortedArr)
	}

	return result
}

func processSingleHandler(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortSequential(reqPayload.ToSort)
	duration := time.Since(startTime)

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       fmt.Sprintf("%dns", duration.Nanoseconds()),
	}

	sendResponse(w, response)
}

func processConcurrentHandler(w http.ResponseWriter, r *http.Request) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := sortConcurrent(reqPayload.ToSort)
	duration := time.Since(startTime)

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       fmt.Sprintf("%dns", duration.Nanoseconds()),
	}

	sendResponse(w, response)
}

func sendResponse(w http.ResponseWriter, response ResponsePayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/process-single", processSingleHandler)
	http.HandleFunc("/process-concurrent", processConcurrentHandler)

	fmt.Println("Server is running on port 8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
