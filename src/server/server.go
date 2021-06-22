package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/logging"
	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
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

// SetController sets the controller used for all functions in the server module
func SetController(controller util.ControllerInterface) {
	cntr = controller
}

// CrawlMiddleware handles some processing of incoming HTTP requests before
// passing on the requests to their specified endpoint
func CrawlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		vars["startTime"] = strconv.FormatInt(startTime, 10)
		// Don't bother with middleware checks
		if _, ok := middlewareBlackList[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		// _, err := DecodeBody(r, vars)
		// if err != nil {
		// 	sendErrorResponse(w, r, http.StatusBadRequest, vars["startTime"], "invalid input")
		// 	return
		// }

		next.ServeHTTP(w, r)
	})
}

func statLookup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	apiKeys, err := util.GetAPIKeys(cntr)
	util.CheckErr(err)

	userStats, err := util.GetUserDetails(cntr, apiKeys[0], vars["steamID0"])
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		res := struct {
			Status int
			Body   string
		}{
			Status: http.StatusInternalServerError,
			Body:   "something went wrong",
		}
		json.NewEncoder(w).Encode(res)
		LogCall(req, http.StatusInternalServerError, vars["startTime"], false)
		logging.SpecialLog(cntr, "errorLog", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userStats)
	LogCall(req, http.StatusOK, vars["startTime"], false)
}

func crawl(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	reqConfig, err := DecodeBody(req, vars)
	if err != nil {
		sendErrorResponse(w, req, http.StatusBadRequest, vars["startTime"], "invalid input")
		return
	}

	apiKeys, err := util.GetAPIKeys(cntr)
	util.CheckErr(err)

	statMode := false
	if reqConfig.StatMode == "true" {
		statMode = true
	}

	level, _ := strconv.Atoi(reqConfig.Level)
	workers, _ := strconv.Atoi(reqConfig.Workers)

	crawlConfig := worker.CrawlerConfig{
		Level:    level,
		StatMode: statMode,
		TestKeys: false,
		Workers:  workers,
		APIKeys:  apiKeys,
	}

	// fmt.Printf("%+v\n", crawlConfig)

	go worker.CrawlOneUser(reqConfig.SteamID0, util.Controller{}, crawlConfig)

	finishedGraphLocation := fmt.Sprintf("%s/%s", configuration.AppConfig.FinishedGraphsLocation, configuration.AppConfig.UrlMap[reqConfig.SteamID0])

	res := struct {
		Body string
	}{
		Body: fmt.Sprintf("Your finished graph will be saved under %s", finishedGraphLocation),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	LogCall(req, http.StatusOK, vars["startTime"], false)
}

func crawlOne(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	finishedGraphLocation := ""

	reqConfig, err := DecodeNewBody(req, vars)
	if err != nil {
		sendErrorResponse(w, req, http.StatusBadRequest, vars["startTime"], err.Error())
		return
	}

	apiKeys, err := util.GetAPIKeys(cntr)
	util.CheckErr(err)

	crawlConfig := worker.CrawlerConfig{
		Level:    reqConfig.Level,
		StatMode: false,
		TestKeys: false,
		Workers:  4,
		APIKeys:  apiKeys,
	}

	go worker.CrawlOneUser(reqConfig.SteamIDs[0], util.Controller{}, crawlConfig)

	time.Sleep(10 * time.Millisecond)
	finishedGraphLocation = fmt.Sprintf("%s/%s", configuration.AppConfig.FinishedGraphsLocation, configuration.AppConfig.UrlMap[reqConfig.SteamIDs[0]])

	res := struct {
		Body string
	}{
		Body: fmt.Sprintf("Your finished graph will be saved under %s", finishedGraphLocation),
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
	jsonObj, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonObj))
	LogCall(req, http.StatusOK, vars["startTime"], false)
}

func home(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, filepath.Join(configuration.AppConfig.StaticDirectoryLocation, "index.html"))
}

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", home).Methods("GET")
	r.HandleFunc("/graph/{id}", serveGraph)
	r.HandleFunc("/crawl", crawl).Methods("POST")
	r.HandleFunc("/statlookup", statLookup).Methods("POST")
	r.HandleFunc("/status", status).Methods("POST")
	r.HandleFunc("/crawlOne", crawlOne).Methods("POST")
	r.Use(CrawlMiddleware)

	return r
}

// RunServer initializes and runs the application as a HTTP server
func RunServer(port string) {
	startTime = time.Now()

	mwBlackList := make(map[string]bool)
	mwBlackList["/"] = true
	mwBlackList["/status"] = true
	middlewareBlackList = mwBlackList

	r := setupRouter()

	fs := http.FileServer(http.Dir(configuration.AppConfig.StaticDirectoryLocation))
	r.PathPrefix("/").Handler(http.StripPrefix("/static", fs))

	log.Printf("Starting web server on http://localhost:%s\n", port)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("127.0.0.1:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())

}
