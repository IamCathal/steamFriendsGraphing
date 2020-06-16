package main

import (
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

	if _, err := os.Stat("userData/"); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.Mkdir("userData/", 0755)
	}

	os.Setenv("testing", "")

	var tests = []testInput{
		{"76561198282036055", os.Getenv("APIKEY"), false},
		{"7656119807862962", os.Getenv("APIKEY"), true},
		{"76561198271948679", os.Getenv("APIKEY"), false},
		{"7656119796028793", os.Getenv("APIKEY"), true},
		{"76561198144084014", os.Getenv("APIKEY"), false},
		{"gibberish", os.Getenv("APIKEY"), true},
	}

	var waitG sync.WaitGroup

	for _, testCase := range tests {
		waitG.Add(1)
		_, err := GetFriends(testCase.steamID, os.Getenv("APIKey"), &waitG)
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
