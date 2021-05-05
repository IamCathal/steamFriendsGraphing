// +build service

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/steamFriendsGraphing/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createValidAPIKEYSFile() *os.File {
	file, err := ioutil.TempFile("", "tempAPIKeys.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("apiKey1\napiKey2\napiKey3")
	file.Seek(0, 0)

	return file
}

// func readAndUnmarshal(res *httptest.ResponseRecorder) (basicResponse, error) {
// 	resJSON := basicResponse{}
// 	response, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return resJSON, err
// 	}
// 	err = json.Unmarshal(response, &resJSON)
// 	if err != nil {
// 		return resJSON, err
// 	}
// 	return resJSON, nil
// }

func TestMain(m *testing.M) {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("BWD", fmt.Sprintf("%s/../", path))

	code := m.Run()

	os.Exit(code)
}

// func initRouter() *mux.Router {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/", HomeHandler).Methods("POST")
// 	r.HandleFunc("/crawl", crawl).Methods("POST")
// 	r.HandleFunc("/statlookup", statLookup).Methods("POST")
// 	r.HandleFunc("/status", status).Methods("POST")
// 	r.Use(CrawlMiddleware)

// 	return r
// }

func TestAPIStatus(t *testing.T) {
	assert.HTTPStatusCode(t, status, "POST", "/status", nil, 200)
	assert.HTTPBodyContains(t, status, "POST", "/status", nil, "operational")
}

func TestStatLookupWithExpectedUserStats(t *testing.T) {
	mockController := &util.MockControllerInterface{}
	SetController(mockController)

	expectedGetPlayerSummaryUser := util.UserStatsStruct{
		Response: util.Response{
			Players: []util.Player{
				{
					Avatarfull:   "expected full avatar url",
					Profileurl:   "expected profile url",
					Profilestate: 1,
					Realname:     "Francis Higgins",
					Steamid:      "76561198076045001",
					Personaname:  "expected persona name",
				},
			},
		},
	}
	expectedUserStats := expectedGetPlayerSummaryUser.Response.Players[0]

	file := createValidAPIKEYSFile()

	mockController.On("OpenFile", mock.AnythingOfType("string")).Return(file, nil)
	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(expectedGetPlayerSummaryUser, nil)

	urlVals, _ := url.ParseQuery("steamID0=testSteamID&statmode=true")
	res := assert.HTTPBody(statLookup, "POST", "/statlookup", urlVals)

	resStruct := util.Player{}
	json.Unmarshal([]byte(res), &resStruct)

	assert.Equal(t, expectedUserStats, resStruct)
}

func TestCrawl(t *testing.T) {
	reqBody := "level=3&testkeys=false&workers=4&steamID0=testSteamID&statmode=true"
	urlVals, _ := url.ParseQuery(reqBody)

	assert.HTTPStatusCode(t, crawl, "POST", "/crawl", urlVals, 200)
	assert.HTTPBodyContains(t, crawl, "POST", "/crawl", urlVals, "Your finished graph will be saved under")
}

// func TestServerRun(t *testing.T) {
// 	// There must be a better way to do this
// 	// but it'll work for now and it doesn't
// 	// mess anything else up
// 	fmt.Printf("\n\n")
// 	go RunServer("8085")
// 	time.Sleep(50 * time.Millisecond)
// 	fmt.Printf("\n\n")
// 	os.Exit(0)
// }
