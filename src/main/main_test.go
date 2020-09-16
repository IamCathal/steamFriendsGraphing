package main

import (
	"os"
	"testing"

	"github.com/steamFriendsGraphing/util"
)

func getAPIKeysForTesting() []string {
	apiKeys := make([]string, 0)

	// When being test on the GitHub actions enviroment
	// it should take keys from from the environment variables
	// rather than the non existent APIKEYS.txt file
	if exists := util.IsEnvVarSet("GITHUBACTIONS"); exists {
		apiKeys = append(apiKeys, os.Getenv("APIKEY"))
		apiKeys = append(apiKeys, os.Getenv("APIKEY1"))
	} else {
		apiKeySlice, err := util.GetAPIKeys()
		util.CheckErr(err)

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

func TestConfigInit(t *testing.T) {
	APIKeys := getAPIKeysForTesting()
	// Regular invocation
	testConfig := config{
		level:    1,
		statMode: false,
		testKeys: false,
		workers:  1,
		steamID:  "76561198282036055",
		APIKeys:  APIKeys,
	}
	testConfig.InitCrawling()

	// StatMode invocation
	testConfig2 := config{
		level:    1,
		statMode: true,
		testKeys: false,
		workers:  1,
		steamID:  "76561198144084014",
		APIKeys:  APIKeys,
	}
	testConfig2.InitCrawling()

	// testKeys invocation
	testConfig3 := config{
		level:    1,
		statMode: false,
		testKeys: true,
		workers:  1,
		steamID:  "76561198144084014",
		APIKeys:  APIKeys,
	}
	testConfig3.InitCrawling()
}
