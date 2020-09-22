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

type config struct {
	Level    string `json:"level"`
	StatMode string `json:"statMode"`
	TestKeys string `json:"testKeys"`
	Workers  string `json:"workers"`
	SteamID  string `json:"steamID"`
}

func DecodeBody(r *http.Request, vars map[string]string) (config, error) {
	inputConfig := config{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputConfig)
	if err != nil {
		return inputConfig, err
	}
	vars["level"] = inputConfig.Level
	vars["statMode"] = inputConfig.StatMode
	vars["testKeys"] = inputConfig.TestKeys
	vars["workers"] = inputConfig.Workers
	vars["steamID"] = inputConfig.SteamID
	return inputConfig, nil
}

func CrawlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		vars["startTime"] = strconv.FormatInt(startTime, 10)

		// Don't bother with middleware checks if it's the root endpoint
		if r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		_, err := DecodeBody(r, vars)
		if err != nil {
			log.Fatal(err)
		}

		next.ServeHTTP(w, r)
	})
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

	configText := fmt.Sprintf("Level: %s - StatMode: %s - TestKeys: %s - Workers: %s - SteamID: %s",
		vars["level"], vars["statmode"], vars["testkeys"], vars["workers"], vars["steamID"])

	res := BasicResponse{
		Status: http.StatusOK,
		Body:   fmt.Sprintf("Crawling with config %s", configText),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req.Method, req.URL.Path, "200", vars["startTime"], false)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/crawl", crawl).Methods("POST")
	r.Use(CrawlMiddleware)

	log.Println("Starting web server on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
