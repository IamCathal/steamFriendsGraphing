package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type basicResponse struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

type statusResponse struct {
	Status string        `json:"status"`
	Uptime time.Duration `json:"uptime"`
}

type newConfig struct {
	Level    string `json:"level"`
	StatMode string `json:"statMode"`
	Workers  string `json:"workers"`
	SteamID0 string `json:"steamID0"`
	SteamID1 string `json:"steamID1"`
}

// inMiddlewareBlackist checks if an endpoint is blacklisted from
// the middleware function i.e no input validation should occur
func inMiddlewareBlacklist(endpoint string) bool {
	_, ok := middlewareBlackList[endpoint]
	if ok {
		return true
	}
	return false
}

// LogCall logs a call to the console with it's details
func LogCall(req *http.Request, status string, startTimeString string, cached bool) {
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
	fmt.Printf("[%s] %s%s %s %s%s%s %dms\n", time.Now().Format("02-Jan-2006 15:04:05"), 
		cacheString, req.Method, req.URL.Path, statusColor, status, "\033[0m", delay)
}

// sendErrorResponse sends an error response
func sendErrorResponse(w http.ResponseWriter, r *http.Request, httpStatus int, startTime, errorString string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	res := basicResponse{
		Status: httpStatus,
		Body:   errorString,
	}
	json.NewEncoder(w).Encode(res)
	LogCall(r, strconv.Itoa(httpStatus), startTime, false)
}

// DecodeBody takes a typical request and assigns the configuration given
func DecodeBody(r *http.Request, vars map[string]string) (newConfig, error) {
	inputConfig := newConfig{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputConfig)
	if err != nil {
		return inputConfig, err
	}

	vars["level"] = inputConfig.Level
	vars["statMode"] = inputConfig.StatMode
	vars["workers"] = inputConfig.Workers
	vars["steamID0"] = inputConfig.SteamID0
	vars["steamID1"] = inputConfig.SteamID1
	return inputConfig, nil
}
