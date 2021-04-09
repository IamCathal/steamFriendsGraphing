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
	// startTime is used keep track of the
	// initialization of this process
	startTime time.Time
	// middlewareBlacklist indicates whether
	// a url is to be ignored by the middleware
	middlewareBlackList map[string]bool
	cntr util.ControllerInterface
)

func setController(controller util.ControllerInterface) {
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

// HomeHandler serves the content for the home page
func HomeHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	vars["startTime"] = strconv.FormatInt(startTime, 10)

	res := basicResponse{
		Status: http.StatusOK,
		Body:   "API is operational",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req, "200", vars["startTime"], false)
}

func statLookup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	apiKeys, err := util.GetAPIKeys()
	util.CheckErr(err)

	resultMap, err := util.GetUserDetails(cntr, apiKeys[0], vars["steamID0"])
	util.CheckErr(err)

	res := basicResponse{
		Status: http.StatusOK,
		Body:   fmt.Sprintf("%+v", resultMap),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req, "200", vars["startTime"], false)
}

func crawl(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	// configText := fmt.Sprintf("Level: %s - StatMode: %s - TestKeys: %s - Workers: %s - SteamID: %s",
	// 	vars["level"], vars["statmode"], vars["testkeys"], vars["workers"], vars["steamID0"])

	res := basicResponse{
		Status: http.StatusOK,
		Body:   fmt.Sprintf("Your finished graph will be saved under %s.html", vars["steamID0"]),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req, "200", vars["startTime"], false)
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
	LogCall(req, "200", vars["startTime"], false)
}

func RunServer(port string) {
	startTime = time.Now()

	mwBlackList := make(map[string]bool)
	mwBlackList["/"] = true
	mwBlackList["/status"] = true
	middlewareBlackList = mwBlackList

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("POST")
	r.HandleFunc("/crawl", crawl).Methods("POST")
	r.HandleFunc("/statlookup", statLookup).Methods("POST")
	r.HandleFunc("/status", status).Methods("POST")
	r.Use(CrawlMiddleware)

	log.Printf("Starting web server on http://localhost:%s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}
