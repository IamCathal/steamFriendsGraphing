package util

import (
	"fmt"
	"os"
	"testing"
)

var (
	expectedGetPlayerSummaryUser UserStatsStruct 
)

type MockInterface struct {}

type testInput struct {
	steamID    string
	APIKey     string
	shouldFail bool
}

func failTest(message string, t *testing.T) {
	failMsg := fmt.Sprintf("%s: %s", t.Name(), message)
	t.Errorf(failMsg)
}

func setupStubs() {
	expectedGetPlayerSummaryUser = UserStatsStruct{
		Response: Response{
			Players: []Player{
				Player {
					Steamid: "76561198076045001",
					Timecreated: 0,
					Personaname: "expected pesrsona name",
				},
			},
		},
	}
}

func (m *MockInterface) CallPlayerSummaryAPI(steamID, apiKey string) (UserStatsStruct, error) {
	return expectedGetPlayerSummaryUser, nil
}

func (m *MockInterface) CallIsAPIKeyValidAPI(apiKey string) string {
	return "valid response"
}

func TestMain(m *testing.M) {
	os.Setenv("testing", "")
	path, err := os.Getwd()
	CheckErr(err)
	os.Setenv("BWD", fmt.Sprintf("%s/../", path))

	setupStubs()
	code := m.Run()
	
	os.Exit(code)
}

func getAPIKeysForTesting() []string {
	apiKeys := make([]string, 0)

	// When being tested on the GitHub actions environment
	// keys should be taken from the environment variables
	// rather than the non existent APIKEYS.txt file
	if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
		apiKeys = append(apiKeys, os.Getenv("APIKEY"))
		apiKeys = append(apiKeys, os.Getenv("APIKEY1"))
	} else {
		apiKeySlice, err := GetAPIKeys()
		CheckErr(err)
		apiKeys = apiKeySlice
	}

	return apiKeys
}

func TestGetPlayerSummary(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	expectedSteamID := expectedGetPlayerSummaryUser.Response.Players[0].Steamid

	cntr := &MockInterface{}
	userDetails, _ := GetPlayerSummary(cntr, expectedSteamID, apiKeys[0])

	if userDetails.Response.Players[0].Steamid != expectedSteamID {
		failMsg := fmt.Sprintf("expected SteamID: %s but received SteamID: %s", expectedSteamID, userDetails.Response.Players[0].Steamid)
		failTest(failMsg, t)
	}
}

func TestCheckAPIKeys(t *testing.T) {
	cntr := &MockInterface{}
	apiKeys := getAPIKeysForTesting()
	ALTCheckAPIKeys(cntr, apiKeys)
}

// Disabled until I can find a way to stub a function twice
// func TestInvalidGetUserDetails(t *testing.T) {
// 	apiKeys := getAPIKeysForTesting()
// 	steamID := "eeee"

// 	cntr := &MockInterface{}
// 	_, err := GetUserDetails(cntr, apiKeys[0], steamID)
// 	if err == nil {
// 		failMsg := fmt.Sprintf("expected error for steamID: %s\n", steamID)
// 		failTest(failMsg, t)
// 	}
// }

func TestGetUsername(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "76561197960287930"

	_, err := GetUsername(&MockInterface{}, apiKeys[0], steamID)
	if err != nil {
		failMsg := fmt.Sprintf("can't get username for %s using key %s", apiKeys[0], steamID)
		failTest(failMsg, t)
	}
}

func TestInvalidGetUsername(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "invalid username"

	_, err := GetUsername(&MockInterface{}, apiKeys[0], steamID)
	if err == nil {
		failMsg := fmt.Sprintf("can't get username for %s using key %s", steamID, apiKeys[0])
		failTest(failMsg, t)
	}
}

func TestCreateDataFolder(t *testing.T) {
	err := CreateUserDataFolder()
	if err != nil {
		failTest("error creating user data folder", t)
	}
	os.RemoveAll("../testData/")
}

func TestIsValidAPIResponseForSteamID(t *testing.T) {
	if isValid := IsValidAPIResponseForSteamId("Internal Server Error"); isValid {
		failTest("Failed to catch invalid steamID", t)
	}

	if isValid := IsValidAPIResponseForSteamId("12 for €8.69 on Galahads is a mad deal"); !isValid {
		failTest("Invalid steamID given for valid response", t)
	}
}

func TestIsValidResponseForAPIKey(t *testing.T) {
	if isValid := IsValidResponseForAPIKey("Forbidden"); isValid {
		failTest("Failed to catch invalid steamID", t)
	}

	if isValid := IsValidResponseForAPIKey("8 for €12 on Heineken is not a mad deal"); !isValid {
		failTest("Invalid steamID given for valid response", t)
	}
}

func TestExtractSteamIDs(t *testing.T) {
	steamIDs := []string{"76561198090461077", "76561198130544932"}
	IDs, err := ExtractSteamIDs(steamIDs)
	if err != nil {
		failTest(err.Error(), t)
	}
	if len(IDs) != 2 {
		failTest("two valid steamIDs given and only 1 returned", t)
	}

	steamIDs2 := []string{}
	_, err = ExtractSteamIDs(steamIDs2)
	if err == nil {
		failTest("no error given for empty steamID slice", t)
	}

}

func TestIsValidFormatSteamID(t *testing.T) {
	if isValid := IsValidFormatSteamID("76561198087169600"); !isValid {
		failMsg := fmt.Sprintf("TestIsValidFormatSteamID: expected %v for steamID: 76561198087169600\n", isValid)
		failTest(failMsg, t)
	}

	if isValid := IsValidFormatSteamID("eeeeeeeee"); isValid {
		failMsg := fmt.Sprintf("TestIsValidFormatSteamID: expected %v for steamID: eeeeeeeee\n", isValid)
		failTest(failMsg, t)
	}
}