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

func statLookup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	apiKeys, err := util.GetAPIKeys()
	util.CheckErr(err)

	resultMap, err := util.GetUserDetails(apiKeys[0], vars["steamID"])
	util.CheckErr(err)

	res := BasicResponse{
		Status: http.StatusOK,
		Body:   fmt.Sprintf("%+v", resultMap),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req.Method, req.URL.Path, "200", vars["startTime"], false)
}

func crawl(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

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

func RunServer(port string) {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/crawl", crawl).Methods("POST")
	r.HandleFunc("/statlookup", statLookup).Methods("POST")
	r.Use(CrawlMiddleware)

	log.Printf("Starting web server on http://localhost:%s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}
