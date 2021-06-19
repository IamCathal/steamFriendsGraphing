// +build service

package server

import (
	"encoding/json"
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

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}

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

	mockController.On("Open", mock.AnythingOfType("string")).Return(file, nil)
	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(expectedGetPlayerSummaryUser, nil)

	urlVals, _ := url.ParseQuery("steamID0=testSteamID&statmode=true")
	res := assert.HTTPBody(statLookup, "POST", "/statlookup", urlVals)

	resStruct := util.Player{}
	json.Unmarshal([]byte(res), &resStruct)

	assert.Equal(t, expectedUserStats, resStruct)
}
