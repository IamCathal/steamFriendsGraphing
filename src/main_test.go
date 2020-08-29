package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
)

type testGetFriendsInput struct {
	steamID    string
	apiKey     string
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

	if _, err := os.Stat("../userData/"); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.Mkdir("../userData/", 0755)
	}

	os.Setenv("testing", "")

	if os.Getenv("APIKEY") == "" {
		t.Error("No APIKEY set")
	}

	apiKeys := getAPIKeysForTesting()

	var tests = []testGetFriendsInput{
		{"76561198282036055", apiKeys[rand.Intn(len(apiKeys))], false},
		{"76561198282036055", "invalid key", true},
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
			t.Error("Caught misbehaving testcase",
				"APIKEY", os.Getenv("APIKEY"),
				"SteamID:", testCase.steamID,
			)
		}
		waitG.Wait()
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
	err := PrintUserDetails(apiKeys[0], steamID)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("")
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
	steamID := "76561198144084014"
	fmt.Println("")
	controlFunc(apiKeys, steamID, true)
	fmt.Println("")

}
