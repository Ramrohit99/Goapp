// main.go
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
	TimeNS       int64   `json:"time_ns"`
}

func main() {
	http.HandleFunc("/process-single", ProcessSingle)
	http.HandleFunc("/process-concurrent", ProcessConcurrent)

	fmt.Println("Server is listening on :8000")
	http.ListenAndServe(":8000", nil)
}

func ProcessSingle(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, sequentialSort)
}

func ProcessConcurrent(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, concurrentSort)
}

func handleRequest(w http.ResponseWriter, r *http.Request, sortFunc func([]int) []int) {
	var reqPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := make([][]int, len(reqPayload.ToSort))
	for i, arr := range reqPayload.ToSort {
		sortedArrays[i] = sortFunc(arr)
	}
	endTime := time.Now()

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       endTime.Sub(startTime).Nanoseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sequentialSort(arr []int) []int {
	sort.Ints(arr)
	return arr
}

func concurrentSort(arr []int) []int {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		sort.Ints(arr)
	}()

	wg.Wait()
	return arr
}
