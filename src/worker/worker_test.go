// +build service

package worker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/logging"
	"github.com/steamFriendsGraphing/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	appConfig = configuration.InitConfig("testing", false)
	// Initialise config for all packages that interact
	// with either log or cache files
	util.SetConfig(appConfig)
	SetConfig(appConfig)
	logging.SetConfig(appConfig)

	code := m.Run()

	os.Exit(code)
}

func TestDivmod(t *testing.T) {
	testNumber := 140

	quotient, remainder := Divmod(testNumber, 100)
	assert.Equal(t, 1, quotient)
	assert.Equal(t, 40, remainder)
}

func TestInitWorkerConfigWithValidInformation(t *testing.T) {
	expectedLevelCap := 2
	expectedWorkerAmount := 40

	workerConfig, err := InitWorkerConfig(expectedLevelCap, expectedWorkerAmount)

	assert.Nil(t, err)
	assert.Equal(t, expectedLevelCap, workerConfig.LevelCap)
	assert.Equal(t, expectedWorkerAmount, workerConfig.WorkerAmount)
}

func TestInitWorkerConfigWithInvalidLevelCap(t *testing.T) {
	expectedLevelCap := -1
	expectedWorkerAmount := 40

	expectedError := errors.New(fmt.Sprintf("invalid level %d given. levelCap must be in range 1-4 (inclusive)", expectedLevelCap))

	workerConfig, err := InitWorkerConfig(expectedLevelCap, expectedWorkerAmount)

	assert.Empty(t, workerConfig)
	assert.EqualError(t, err, expectedError.Error())
}

func TestInitWorkerConfigWithInvalidWorkerAmount(t *testing.T) {
	expectedLevelCap := 2
	expectedWorkerAmount := 129

	expectedError := errors.New(fmt.Sprintf("invalid worker amount %d given. worker amount must be in range 1-60 (inclusive)", expectedLevelCap))

	workerConfig, err := InitWorkerConfig(expectedLevelCap, expectedWorkerAmount)

	assert.Empty(t, workerConfig)
	assert.EqualError(t, err, expectedError.Error())
}
func TestGetFriendsWithValidInformation(t *testing.T) {
	mockController := &util.MockControllerInterface{}
	originalUserSteamID := "76561198282036055"

	eddieDurcanSteamID := "007"
	eddieDurcanUsername := "eddieDurcan247"
	eddieDurcanFriendsSince := 5

	frenchToastSteamID := "008"
	frenchToastUsername := "toasteen"
	frenchToastFriendsSince := 8

	// Create a folder to hold the logfile generated
	os.Mkdir(appConfig.LogsFolderLocation, 0755)

	apiKeys := []string{"apiKey1", "apiKey2"}
	jobs := make(chan JobsStruct, 100)

	testCase := struct {
		steamID  string
		apikey   string
		statMode bool
	}{
		originalUserSteamID,
		apiKeys[rand.Intn(len(apiKeys))],
		false,
	}

	friendsInfoForOriginalUser := util.FriendsStruct{
		FriendsList: util.Friendslist{
			Friends: []util.Friend{
				{
					Steamid:      eddieDurcanSteamID,
					Relationship: "friend",
					FriendSince:  eddieDurcanFriendsSince,
				},
				{
					Steamid:      frenchToastSteamID,
					Relationship: "friend",
					FriendSince:  frenchToastFriendsSince,
				},
			},
		},
	}

	friendsInfoForOriginalUserUserStats := util.UserStatsStruct{
		Response: util.Response{
			Players: []util.Player{
				{
					Steamid:     eddieDurcanSteamID,
					Personaname: eddieDurcanUsername,
				},
				{
					Steamid:     frenchToastSteamID,
					Personaname: frenchToastUsername,
				},
			},
		},
	}

	friendsUsernamesForOriginalUser := util.FriendsStruct{
		FriendsList: util.Friendslist{
			Friends: []util.Friend{
				{
					Steamid:      eddieDurcanSteamID,
					Relationship: "friend",
					Username:     eddieDurcanUsername,
					FriendSince:  eddieDurcanFriendsSince,
				},
				{
					Steamid:      frenchToastSteamID,
					Relationship: "friend",
					Username:     frenchToastUsername,
					FriendSince:  frenchToastFriendsSince,
				},
			},
		},
	}

	os.Setenv("CURRTARGET", testCase.steamID)

	// A real file pointer must be used. Creating a nil pointer os.File
	// results in an ErrInvalid when closed
	dummyFile, err := ioutil.TempFile("", "tempAPIKeys.txt")
	if err != nil {
		log.Fatal(err)
	}
	mockController.On("CreateFile", mock.AnythingOfType("string")).Return(dummyFile, nil)
	mockController.On("WriteGzip", mock.AnythingOfType("*os.File"), mock.AnythingOfType("string")).Return(nil)

	expectedLogsFile := fmt.Sprintf("%s/%s.txt", appConfig.LogsFolderLocation, testCase.steamID)
	tempLogFile, err := os.Create(expectedLogsFile)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tempLogFile.Name())
	mockController.On("OpenFile", expectedLogsFile, mock.Anything, mock.Anything).Return(tempLogFile, nil)

	mockController.On("FileExists", mock.AnythingOfType("string")).Return(false)
	mockController.On("CallGetFriendsListAPI", originalUserSteamID, mock.AnythingOfType("string")).Return(friendsInfoForOriginalUser, nil)

	// Used to get the friendslist of the target user
	mockController.On("CallPlayerSummaryAPI", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(friendsInfoForOriginalUserUserStats, nil)
	// Used to get the username of the current target user
	mockController.On("CallPlayerSummary", originalUserSteamID, mock.AnythingOfType("string")).Return(friendsUsernamesForOriginalUser, nil)

	friends, err := GetFriends(mockController, testCase.steamID, testCase.apikey, 1, jobs)

	assert.Nil(t, err)
	assert.Equal(t, friendsUsernamesForOriginalUser.FriendsList, friends.FriendsList)

	os.RemoveAll(appConfig.LogsFolderLocation)
}

func TestGetFriendsWithInvalidGetFriendsAPICallWhenRetrievingTargetUsersFriends(t *testing.T) {
	mockController := &util.MockControllerInterface{}
	originalUserSteamID := "76561198282036055"

	apiKeys := []string{"apiKey1", "apiKey2"}
	jobs := make(chan JobsStruct, 100)

	testCase := struct {
		steamID  string
		apikey   string
		statMode bool
	}{
		originalUserSteamID,
		apiKeys[rand.Intn(len(apiKeys))],
		false,
	}

	// Create a folder to hold the logfile generated
	os.Mkdir(appConfig.LogsFolderLocation, 0755)

	os.Setenv("CURRTARGET", testCase.steamID)
	mockController.On("FileExists", mock.AnythingOfType("string")).Return(false)

	expectedLogsFile := fmt.Sprintf("%s/%s.txt", appConfig.LogsFolderLocation, testCase.steamID)
	tempLogFile, err := os.Create(expectedLogsFile)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tempLogFile.Name())
	mockController.On("OpenFile", expectedLogsFile, mock.AnythingOfType("int"), mock.AnythingOfType("os.FileMode")).Return(tempLogFile, nil)

	getFriendsListAPIError := errors.New("error")
	mockController.On("CallGetFriendsListAPI", originalUserSteamID, mock.AnythingOfType("string")).Return(util.FriendsStruct{}, getFriendsListAPIError)

	friends, err := GetFriends(mockController, testCase.steamID, testCase.apikey, 1, jobs)

	assert.Empty(t, friends)
	assert.EqualError(t, err, getFriendsListAPIError.Error())

	os.RemoveAll(appConfig.LogsFolderLocation)
}

func TestGetFriendsWithInvalidFormatSteamID(t *testing.T) {
	mockController := &util.MockControllerInterface{}
	originalUserSteamID := "invalid"

	apiKeys := []string{"apiKey1", "apiKey2"}
	jobs := make(chan JobsStruct, 100)

	testCase := struct {
		steamID  string
		apikey   string
		statMode bool
	}{
		originalUserSteamID,
		apiKeys[rand.Intn(len(apiKeys))],
		false,
	}

	// Create a folder to hold the logfile generated
	os.Mkdir(appConfig.LogsFolderLocation, 0755)

	expectedLogsFile := fmt.Sprintf("%s/%s.txt", appConfig.LogsFolderLocation, testCase.steamID)
	tempLogFile, err := os.Create(expectedLogsFile)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tempLogFile.Name())
	mockController.On("OpenFile", expectedLogsFile, mock.AnythingOfType("int"), mock.AnythingOfType("os.FileMode")).Return(tempLogFile, nil)

	os.Setenv("CURRTARGET", testCase.steamID)
	mockController.On("FileExists", mock.AnythingOfType("string")).Return(false)
	expectedError := errors.New(fmt.Sprintf("invalid steamID: %s, apikey: %s", testCase.steamID, testCase.apikey))

	friends, err := GetFriends(mockController, testCase.steamID, testCase.apikey, 1, jobs)

	assert.Empty(t, friends)
	assert.Contains(t, err.Error(), expectedError.Error())

	os.RemoveAll(appConfig.LogsFolderLocation)
}

func TestIsEnvVarSetWithValidEnvVar(t *testing.T) {
	os.Setenv("examplevariable", "thisIsSet")
	exists := IsEnvVarSet("examplevariable")

	assert.True(t, exists)
}

func TestIsEnvVarSetWithInvalidEnvVar(t *testing.T) {
	exists := IsEnvVarSet("nonexistantvariable")

	assert.False(t, exists)
}
