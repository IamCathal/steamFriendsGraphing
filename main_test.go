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
		{"76561197999662696", os.Getenv("APIKey"), false},
		{"7656119807862962", os.Getenv("APIKey"), true},
		{"76561197960435530", os.Getenv("APIKey"), false},
		{"7656119796028793", os.Getenv("APIKey"), true},
		{"76561198023414915", os.Getenv("APIKey"), false},
		{"gibberish", os.Getenv("APIKey"), true},
	}

	for _, testCase := range tests {
		_, err := GetFriends(testCase.steamID, os.Getenv("APIKey"))
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
