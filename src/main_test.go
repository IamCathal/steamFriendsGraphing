package main

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

type testGetFriendsInput struct {
	steamID    string
	apiKey     string
	shouldFail bool
}

type testControlFuncInput struct {
	steamID  string
	statMode bool
}

type testGetUsernameFromCacheFile struct {
	steamID    string
	shouldFail bool
}

func getAPIKeysForTesting() []string {
	apiKeys := make([]string, 0)

	// When being test on the GitHub actions enviroment
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

// Setup and teardown function
func TestMain(m *testing.M) {
	os.Setenv("testing", "")
	if _, err := os.Stat("../testData/"); os.IsNotExist(err) {
		os.Mkdir("../testData/", 0755)
	}

	code := m.Run()

	os.RemoveAll("../testData")
	os.Exit(code)
}

func TestAllAPIKeys(t *testing.T) {
	apiKeys := getAPIKeysForTesting()

	CheckAPIKeys(apiKeys)
}

func TestExampleInvocation(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	fmt.Printf("==================== Example Invocation ====================\n")
	newControlFunc(apiKeys, "76561198130544932", 2)
	fmt.Printf("\n")
	newControlFunc(apiKeys, "76561198130544932", 1)
	fmt.Printf("\n")
	newControlFunc(apiKeys, "76561198130544932", 1)
	fmt.Printf("============================================================\n")
}

func TestGetFriends(t *testing.T) {
	apiKeys := getAPIKeysForTesting()

	var tests = []testGetFriendsInput{
		{"76561198282036055", apiKeys[rand.Intn(len(apiKeys))], false},
		{"76561198081485934", "invalid key", true},
		{"7656119807862962", apiKeys[rand.Intn(len(apiKeys))], true},
		{"76561198271948679", apiKeys[rand.Intn(len(apiKeys))], false},
		{"7656119796028793", apiKeys[rand.Intn(len(apiKeys))], true},
		{"76561198144084014", apiKeys[rand.Intn(len(apiKeys))], false},
		{"11111111111111111", apiKeys[rand.Intn(len(apiKeys))], true},
		{"gibberish", apiKeys[rand.Intn(len(apiKeys))], true},
	}

	for _, testCase := range tests {
		_, err := GetFriends(testCase.steamID, testCase.apiKey)
		if err != nil {
			if !testCase.shouldFail {
				t.Error("Error:", err,
					"SteamID:", testCase.steamID,
				)
			}
		} else if testCase.shouldFail {
			t.Error("caught misbehaving testcase",
				"APIKEY:", testCase.apiKey,
				"SteamID:", testCase.steamID,
			)
		}
	}
}

func TestCreateDataFolder(t *testing.T) {
	err := CreateUserDataFolder()
	if err != nil {
		t.Error("error creating user data folder")
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

func TestInvalidUsername(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "invalid username"

	_, err := GetUsername(apiKeys[0], steamID)
	if err == nil {
		t.Errorf("can't get username for %s using key %s", steamID, apiKeys[0])
	}
}

func TestPrintUserDetails(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "76561197960287930"

	fmt.Println("")
	fmt.Printf("==================== Print user details ====================\n")
	err := PrintUserDetails(apiKeys[0], steamID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("============================================================\n\n")
}

func TestInvalidPrintUserDetails(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	steamID := "bad input"

	err := PrintUserDetails(apiKeys[0], steamID)
	if err == nil {
		t.Error(err)
	}
}

func TestGetUsernameFromCacheFile(t *testing.T) {

	var tests = []testGetUsernameFromCacheFile{
		{"76561198306145504", true},
		{"76561198096639661", false},
	}
	for _, elem := range tests {
		_, err := GetUsernameFromCacheFile(elem.steamID)
		if err != nil {
			if !elem.shouldFail {
				t.Errorf("didn't fail to get username for %s", elem.steamID)
			}
		} else if elem.shouldFail {
			t.Error("caught misbehaving testcase",
				"SteamID:", elem.steamID,
			)
		}
	}
}

func TestGetCache(t *testing.T) {
	_, err := GetCache("76561198282036055e")
	if err == nil {
		t.Error("invalid cache get did not throw an error")
	}

	// Will fail if not run as part of the whole test suite
	// as this cache file will not have been written
	_, err = GetCache("76561198245030292")
	if err == nil {
		t.Error("Got invalid cache for user 76561198245030292")
	}
}

func TestInitWorkerConfig(t *testing.T) {
	_, err := InitWorkerConfig(4)
	if err != nil {
		t.Error(err)
	}

	_, err = InitWorkerConfig(-2)
	if err == nil {
		t.Errorf("failed to catch invalid levelCap of -2")
	}
}
