package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type statusResponse struct {
	Status string        `json:"status"`
	Uptime time.Duration `json:"uptime"`
}

type requestConfig struct {
	Level    int      `json:"level"`
	SteamIDs []string `json:"steamIDs"`
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
func LogCall(req *http.Request, status int, startTimeString string, cached bool) {
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
	if status < 400 && status > 199 {
		statusColor = "\033[32m"
	} else {
		statusColor = "\033[31m"
	}
	fmt.Printf("[%s] %s%s %s%d%s %s %s %s %dms\n", time.Now().Format("02-Jan-2006 15:04:05"),
		cacheString, req.Method, statusColor, status, "\033[0m", req.URL.Path, req.RemoteAddr, req.UserAgent(), delay)
}

// sendErrorResponse sends an error response
func sendErrorResponse(w http.ResponseWriter, r *http.Request, httpStatus int, startTime, errorString string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	res := struct {
		Response string `json:"response"`
	}{
		Response: errorString,
	}
	json.NewEncoder(w).Encode(res)
	LogCall(r, httpStatus, startTime, false)
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

// DecodeBody takes a typical request and assigns the configuration given
func DecodeNewBody(r *http.Request, vars map[string]string) (requestConfig, error) {
	reqConfig := requestConfig{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqConfig)
	if err != nil {
		return requestConfig{}, err
	}

	vars["level"] = strconv.Itoa(reqConfig.Level)

	switch len(reqConfig.SteamIDs) {
	case 1:
		vars["steamID0"] = reqConfig.SteamIDs[0]
	case 2:
		vars["steamID1"] = reqConfig.SteamIDs[1]
	default:
		return requestConfig{}, errors.New("invalid amount of steamIDs given")
	}

	return reqConfig, nil
}
