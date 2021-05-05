package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/steamFriendsGraphing/util"
)

var (
	cntr util.ControllerInterface
	// startTime is used keep track of the
	// initialization of this process
	startTime time.Time
	// middlewareBlacklist indicates whether
	// a url is to be ignored by the middleware
	middlewareBlackList map[string]bool
)

func SetController(controller util.ControllerInterface) {
	cntr = controller
}

func CrawlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		vars["startTime"] = strconv.FormatInt(startTime, 10)
		// Don't bother with middleware checks if it's the root endpoint
		if r.URL.Path == "/" || r.URL.Path == "/status" {
			next.ServeHTTP(w, r)
			return
		}

		_, err := DecodeBody(r, vars)
		if err != nil {
			sendErrorResponse(w, r, http.StatusBadRequest, vars["startTime"], "invalid input")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func statLookup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	apiKeys, err := util.GetAPIKeys(cntr)
	util.CheckErr(err)

	userStats, err := util.GetUserDetails(cntr, apiKeys[0], vars["steamID0"])
	util.CheckErr(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userStats)
	LogCall(req, http.StatusOK, vars["startTime"], false)
}

func crawl(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	// configText := fmt.Sprintf("Level: %s - StatMode: %s - TestKeys: %s - Workers: %s - SteamID: %s",
	// 	vars["level"], vars["statmode"], vars["testkeys"], vars["workers"], vars["steamID0"])

	res := struct {
		Body string
	}{
		Body: fmt.Sprintf("Your finished graph will be saved under %s.html", vars["steamID0"]),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req, http.StatusOK, vars["startTime"], false)
}

func status(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	res := statusResponse{
		Uptime: time.Since(startTime),
		Status: "operational",
	}
	jsonObj, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonObj))
	LogCall(req, http.StatusOK, vars["startTime"], false)
}

func RunServer(port string) {
	startTime = time.Now()
	SetController(util.Controller{})

	mwBlackList := make(map[string]bool)
	mwBlackList["/"] = true
	mwBlackList["/status"] = true
	middlewareBlackList = mwBlackList

	r := mux.NewRouter()
	r.HandleFunc("/crawl", crawl).Methods("POST")
	r.HandleFunc("/statlookup", statLookup).Methods("POST")
	r.HandleFunc("/status", status).Methods("POST")
	r.Use(CrawlMiddleware)

	log.Printf("Starting web server on http://localhost:%s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}
