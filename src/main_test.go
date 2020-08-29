package main

import (
	"math/rand"
	"os"
	"sync"
	"testing"
)

type testInput struct {
	steamID    string
	apiKey     string
	shouldFail bool
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

	apiKeys, err := GetAPIKeys()
	CheckErr(err)

	var tests = []testInput{
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
