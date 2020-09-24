package graphing

import (
	"os"
	"testing"

	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
)

func getAPIKeysForTesting() []string {
	apiKeys := make([]string, 0)

	// When being test on the GitHub actions environment
	// it should take keys from from the environment variables
	// rather than the non existent APIKEYS.txt file
	if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
		apiKeys = append(apiKeys, os.Getenv("APIKEY"))
		apiKeys = append(apiKeys, os.Getenv("APIKEY1"))
	} else {
		apiKeySlice, err := util.GetAPIKeys()
		util.CheckErr(err)

		apiKeys = apiKeySlice
	}

	return apiKeys
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

func TestGraphing(t *testing.T) {
	APIKeys := getAPIKeysForTesting()

	// Must first fetch the data, otherwise there would
	// be no cached files to construct the graph with
	testConfig := worker.CrawlerConfig{
		Level:    2,
		StatMode: false,
		TestKeys: false,
		Workers:  1,
		APIKeys:  APIKeys,
	}
	worker.InitCrawling(testConfig, "76561198090461077")

	InitGraphing(2, 2, "76561198090461077")
}
