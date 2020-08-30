package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
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
	if os.Getenv("GITHUBACTIONS") != "" {
		apiKeys = append(apiKeys, os.Getenv("APIKEY"))
		apiKeys = append(apiKeys, os.Getenv("APIKEY1"))
	} else {
		apiKeySlice, err := GetAPIKeys()
		CheckErr(err)

		apiKeys = apiKeySlice
	}

	return apiKeys
}

func TestGetFriends(t *testing.T) {

	if _, err := os.Stat("../testData/"); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.Mkdir("../testData/", 0755)
	}

	os.Setenv("testing", "")
	os.Setenv("disablereadcache", "")

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

	var waitG sync.WaitGroup

	for _, testCase := range tests {
		waitG.Add(1)
		_, err := GetFriends(testCase.steamID, testCase.apiKey, &waitG)
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
		waitG.Wait()
	}
	// Add a space and sleep for a second to keep
	// the logs nice and clean
	time.Sleep(2 * time.Second)
	fmt.Printf("\n")

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

func TestExampleInvocation(t *testing.T) {
	apiKeys := getAPIKeysForTesting()

	var tests = []testControlFuncInput{
		{"76561198282036055", true},
		{"76561198210804275", true},
		{"76561198245030292", false},
	}

	fmt.Printf("\n================== Test Example Invocation =================\n")
	for _, elem := range tests {
		controlFunc(apiKeys, elem.steamID, elem.statMode)
	}
	time.Sleep(2 * time.Second)
	fmt.Printf("============================================================\n\n")
}

func TestGetUsernameFromCacheFile(t *testing.T) {

	var tests = []testGetUsernameFromCacheFile{
		{"76561198306145504", true},
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

func TestCleanup(t *testing.T) {
	os.RemoveAll("../testData")
}
