package main

import (
	"log"
	"testing"
)

type testInput struct {
	steamID    string
	apiKey     string
	shouldFail bool
}

func TestGetFriends(t *testing.T) {
	apiKeys, err := GetAPIKeys()
	if err != nil {
		log.Fatal(err)
	}

	var tests = []testInput{
		{"76561198078629620", apiKeys[0], false},
		{"7656119807862962", apiKeys[1], true},
		{"76561197960435530", apiKeys[2], false},
		{"7656119796028793", apiKeys[3], true},
		{"76561198023414915", apiKeys[4], false},
		{"gibberish", apiKeys[5], true},
	}

	for i, testCase := range tests {
		_, err := GetFriends(testCase.steamID, apiKeys[i])
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
