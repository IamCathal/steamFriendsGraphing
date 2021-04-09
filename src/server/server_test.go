package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/steamFriendsGraphing/util"
)

var (
	expectedGetPlayerSummaryUser util.UserStatsStruct 
)

type MockInterface struct {}

func setupStubs() {
	expectedGetPlayerSummaryUser = util.UserStatsStruct{
		Response: util.Response{
			Players: []util.Player{
				util.Player {
					Steamid: "76561198076045001",
					Timecreated: 0,
					Personaname: "expected pesrsona name",
				},
			},
		},
	}
}

func (m *MockInterface) CallPlayerSummaryAPI(steamID, apiKey string) (util.UserStatsStruct, error) {
	return expectedGetPlayerSummaryUser, nil
}

func (m *MockInterface) CallIsAPIKeyValidAPI(apiKey string) string {
	return "valid response"
}

func failTest(message string, t *testing.T) {
	failMsg := fmt.Sprintf("%s: %s", t.Name(), message)
	t.Errorf(failMsg)
}

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

func TestMain(m *testing.M) {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("BWD", fmt.Sprintf("%s/../", path))

	setupStubs()

	code := m.Run()
	os.Exit(code)
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
		failTest("Error reading response", t)
	}

	if resJSON.Status != 200 || resJSON.Body != "API is operational" {
		failTest("Root endpoint not operational", t)
	}
}

func TestStatLookup(t *testing.T) {
	cntr := &MockInterface{}
	setController(cntr)
	reqBody, err := json.Marshal(map[string]string{
		"steamID0": "76561197960271945",
		"statMode": "true",
	})
	if err != nil {
		failTest(err.Error(), t)
	}

	req, _ := http.NewRequest("POST", "/statlookup", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()
	initRouter().ServeHTTP(res, req)

	resJSON, err := readAndUnmarshal(res)
	if err != nil {
		failTest("Error unmarshaling response", t)
	}

	if resJSON.Status != 200 {
		failTest("expect 200 respons", t)
	}
}

func TestCrawl(t *testing.T) {
	reqBody, err := json.Marshal(map[string]string{
		"level":    "3",
		"testkeys": "false",
		"workers":  "4",
		"steamID0": "76561197960271945",
		"statmode": "true",
	})
	if err != nil {
		failTest(err.Error(), t)
	}

	req, _ := http.NewRequest("POST", "/crawl", bytes.NewBuffer(reqBody))
	res := httptest.NewRecorder()
	initRouter().ServeHTTP(res, req)

	resJSON, err := readAndUnmarshal(res)
	if err != nil {
		failTest("Error reading response",t)
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
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("\n\n")
	os.Exit(0)
}
