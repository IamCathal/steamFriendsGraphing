package main

import (
	"os"
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

	var tests = []testInput{
		{"76561197999662696", apiKey, false},
		{"7656119807862962", apiKey, true},
		{"76561197960435530", apiKey, false},
		{"7656119796028793", apiKey, true},
		{"76561198023414915", apiKey, false},
		{"gibberish", apiKey, true},
	}

	for _, testCase := range tests {
		_, err := GetFriends(testCase.steamID, apiKey)
		// fmt.Println(testCase.steamID, testCase.shouldFail, err)
		if err != nil {
			if !testCase.shouldFail {
				t.Error("Error:", err,
					"SteamID:", testCase.steamID,
				)
			}
		} else if testCase.shouldFail {
			t.Error("Caught misbehaving testcase",
				"SteamID:", testCase.steamID,
			)
		}
	}
}
