package util

import (
	"fmt"
	"os"
	"testing"
)

type testInput struct {
	steamID    string
	APIKey     string
	shouldFail bool
}

func TestMain(m *testing.M) {
	os.Setenv("testing", "")
	path, err := os.Getwd()
	CheckErr(err)
	os.Setenv("BWD", fmt.Sprintf("%s/../", path))

	fmt.Println(":: RemoveAll ../testData")
	os.RemoveAll("../testData")
	fmt.Println(":: Mkdir ../testData")
	os.Mkdir("../testData", 0755)

	code := m.Run()
	fmt.Println(":: RemoveAll ../testData")
	os.RemoveAll("../testData")
	os.Exit(code)
}

func getAPIKeysForTesting() []string {
	apiKeys := make([]string, 0)

	// When being tested on the GitHub actions environment
	// it should take keys from from the environment variables
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

func TestAllAPIKeys(t *testing.T) {
	apiKeys := getAPIKeysForTesting()

	CheckAPIKeys(apiKeys)
}

func TestValidGetUserDetails(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "76561198076045001"

	expectedDetails := make(map[string]string, 0)
	expectedDetails["SteamID"] = steamID
	expectedDetails["TimeCreated"] = "2012-11-18 06:49:56 +0000 GMT"

	userDetails, _ := GetUserDetails(apiKeys[0], steamID)
	if userDetails["SteamID"] != expectedDetails["SteamID"] {
		t.Errorf("TestInvalidGetUserDetails: expected SteamID: %s but received SteamID: %s", expectedDetails["SteamID"], userDetails["SteamID"])
	}

	if userDetails["TimeCreated"] != expectedDetails["TimeCreated"] {
		t.Errorf("TestInvalidGetUserDetails: expected TimeCreated: %s but received TimeCreated: %s", expectedDetails["TimeCreated"], userDetails["TimeCreated"])
	}
}

func TestInvalidGetUserDetails(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "eeee"

	_, err := GetUserDetails(apiKeys[0], steamID)
	if err == nil {
		t.Errorf("TestInvalidGetUserDetails: expected error for steamID: %s\n", steamID)
	}
}

func TestGetUsername(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "76561197960287930"

	_, err := GetUsername(apiKeys[0], steamID)
	if err != nil {
		t.Errorf("can't get username for %s using key %s", apiKeys[0], steamID)
	}
}

func TestInvalidGetUsername(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "invalid username"

	_, err := GetUsername(apiKeys[0], steamID)
	if err == nil {
		t.Errorf("can't get username for %s using key %s", steamID, apiKeys[0])
	}
}


func TestCreateDataFolder(t *testing.T) {
	err := CreateUserDataFolder()
	if err != nil {
		t.Error("error creating user data folder")
	}
	os.RemoveAll("userData/")

}

func TestIsValidSteamID(t *testing.T) {
	if isValid := IsValidSteamID("Internal Server Error"); isValid {
		t.Error("Failed to catch invalid steamID")
	}

	if isValid := IsValidSteamID("12 for €8.69 on Galahads is a mad deal"); !isValid {
		t.Error("Invalid steamID given for valid response")
	}
}

func TestIsAPIKey(t *testing.T) {
	if isValid := IsValidAPIKey("Forbidden"); isValid {
		t.Error("Failed to catch invalid steamID")
	}

	if isValid := IsValidAPIKey("8 for €12 on Heineken is not a mad deal"); !isValid {
		t.Error("Invalid steamID given for valid response")
	}
}

func TestExtractSteamIDs(t *testing.T) {
	steamIDs := []string{"76561198090461077", "76561198130544932"}
	IDs, err := ExtractSteamIDs(steamIDs)
	if err != nil {
		t.Error(err)
	}
	if len(IDs) != 2 {
		t.Error("two valid steamIDs given and only 1 returned")
	}

	steamIDs2 := []string{}
	IDs, err = ExtractSteamIDs(steamIDs2)
	if err == nil {
		t.Errorf("no error given for empty steamID slice")
	}

}

func TestIsValidFormatSteamID(t *testing.T) {
	if isValid := IsValidFormatSteamID("76561198087169600"); !isValid {
		t.Errorf("TestIsValidFormatSteamID: expected %v for steamID: 76561198087169600\n", isValid)
	}

	if isValid := IsValidFormatSteamID("eeeeeeeee"); isValid {
		t.Errorf("TestIsValidFormatSteamID: expected %v for steamID: eeeeeeeee\n", isValid)
	}
}