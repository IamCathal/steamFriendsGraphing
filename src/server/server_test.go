package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func readAndUnmarshal(res *httptest.ResponseRecorder) (basicResponse, error) {
	resJSON := basicResponse{}
	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resJSON, err
	}
	err = json.Unmarshal(response, &resJSON)
	if err != nil {
		return resJSON, err
	}
	return resJSON, nil
}

func initRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("POST")
	r.HandleFunc("/crawl", crawl).Methods("POST")
	r.HandleFunc("/statlookup", statLookup).Methods("POST")
	r.HandleFunc("/status", status).Methods("POST")
	r.Use(CrawlMiddleware)

	return r
}

func TestAPIStatus(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", nil)
	res := httptest.NewRecorder()
	initRouter().ServeHTTP(res, req)

	resJSON, err := readAndUnmarshal(res)
	if err != nil {
		t.Error("Error reading response")
	}

	if resJSON.Status != 200 || resJSON.Body != "API is operational" {
		t.Errorf("Root endpoint not operational")
	}
}

func TestStatLookup(t *testing.T) {
	reqBody, err := json.Marshal(map[string]string{
		"steamID":  "76561197960271945",
		"statMode": "true",
	})
	if err != nil {
		t.Error(err)
	}

	req, _ := http.NewRequest("POST", "/statlookup", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()
	initRouter().ServeHTTP(res, req)

	resJSON, err := readAndUnmarshal(res)
	if err != nil {
		t.Error("Error reading response")
	}

	if resJSON.Status != 200 {
		t.Errorf("Statlookup endpoint not operational")
	}
}

func TestCrawl(t *testing.T) {
	reqBody, err := json.Marshal(map[string]string{
		"level":    "3",
		"testkeys": "false",
		"workers":  "4",
		"steamID":  "76561197960271945",
		"statmode": "true",
	})
	if err != nil {
		t.Error(err)
	}

	req, _ := http.NewRequest("POST", "/crawl", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()
	initRouter().ServeHTTP(res, req)

	resJSON, err := readAndUnmarshal(res)
	if err != nil {
		t.Error("Error reading response")
	}

	if resJSON.Status != 200 {
		t.Errorf("Crawl endpoint not operational")
	}
}

func TestServerRun(t *testing.T) {
	// There must be a better way to do this
	// but it'll work for now and it doesn't
	// mess anything else up
	fmt.Printf("\n\n")
	go RunServer("8085")
	time.Sleep(2 * time.Second)
	fmt.Printf("\n\n")
	os.Exit(0)
}
