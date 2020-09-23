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
	os.RemoveAll("../testData")
	if _, err := os.Stat("../testData"); os.IsNotExist(err) {
		os.Mkdir("../testData", 0755)
	}

	code := m.Run()

	os.RemoveAll("../testData")
	os.Exit(code)
}

func getAPIKeysForTesting() []string {
	apiKeys := make([]string, 0)

	// When being test on the GitHub actions environment
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

func TestInvalidGetUserDetails(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "bad input"

	_, err := GetUserDetails(apiKeys[0], steamID)
	if err == nil {
		t.Error(err)
	}
}

func TestGetUserDetails(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "76561197960287930"

	fmt.Println("")
	fmt.Printf("==================== Print user details ====================\n")
	resMap, err := GetUserDetails(apiKeys[0], steamID)
	if err != nil {
		t.Error(err)
	}
	for k, v := range resMap {
		fmt.Printf("%13s: %s\n", k, v)
	}
	fmt.Printf("============================================================\n\n")
}

func TestInvalidUsername(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "invalid username"

	_, err := GetUsername(apiKeys[0], steamID)
	if err == nil {
		t.Errorf("can't get username for %s using key %s", steamID, apiKeys[0])
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
