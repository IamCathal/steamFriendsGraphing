package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type BasicResponse struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// Log an API call to the console with it's details
func LogCall(method, endpoint, status, startTimeString string, cached bool) {
	statusColor := "\033[0m"
	cacheString := ""

	if cached {
		cacheString = "[CACHE] "
	}

	startTime, err := strconv.ParseInt(startTimeString, 10, 64)
	if err != nil {
		startTime = -1
	}
	endTime := time.Now().UnixNano() / int64(time.Millisecond)
	delay := endTime - startTime

	// If the HTTP status given is 2XX, give it a nice
	// green color, otherwise give it a red color
	if status[0] == '2' {
		statusColor = "\033[32m"
	} else {
		statusColor = "\033[31m"
	}
	fmt.Printf("[%s] %s%s %s %s%s%s %dms\n", time.Now().Format("02-Jan-2006 15:04:05"), cacheString, method, endpoint, statusColor, status, "\033[0m", delay)
}

// HomeHandler serves the content for the home page
func HomeHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	vars["startTime"] = strconv.FormatInt(startTime, 10)

	res := BasicResponse{
		Status: http.StatusOK,
		Body:   "API is operational",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req.Method, req.URL.Path, "200", vars["startTime"], false)
}

func crawl(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	vars["startTime"] = strconv.FormatInt(startTime, 10)

	res := BasicResponse{
		Status: http.StatusOK,
		Body:   fmt.Sprintf("Crawling %s", vars["username"]),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req.Method, req.URL.Path, "200", vars["startTime"], false)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/crawl/{username}", crawl).Methods("POST")

	log.Println("Starting web server on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
