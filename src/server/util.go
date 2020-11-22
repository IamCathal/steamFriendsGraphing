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

type Config struct {
	Level    string `json:"level"`
	StatMode string `json:"statMode"`
	TestKeys string `json:"testKeys"`
	Workers  string `json:"workers"`
	SteamID  string `json:"steamID"`
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

// DecodeBody takes a typical request and assigns the configuration given
func DecodeBody(r *http.Request, vars map[string]string) (Config, error) {
	inputConfig := Config{}
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
