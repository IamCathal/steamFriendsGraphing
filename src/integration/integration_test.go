// +build integration

package integration

import (
	"os"
	"testing"

	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/graphing"
	"github.com/steamFriendsGraphing/logging"
	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
	"github.com/stretchr/testify/assert"
)

var (
	config  configuration.Info
	apiKeys []string
)

func getAPIKeysForTesting() []string {
	cntr := util.Controller{}
	apiKeys := make([]string, 0)

	// When being test on the GitHub actions environment
	// it should take keys from from the environment variables
	// rather than the non existent APIKEYS.txt file
	if exists := worker.IsEnvVarSet("GITHUBACTIONS"); exists {
		apiKeys = append(apiKeys, os.Getenv("APIKEY"))
		apiKeys = append(apiKeys, os.Getenv("APIKEY1"))
	} else {
		apiKeySlice, err := util.GetAPIKeys(cntr)
		util.CheckErr(err)

		apiKeys = apiKeySlice
	}

	return apiKeys
}

func TestMain(m *testing.M) {
	// Setup apikeys and config or all tests
	config := configuration.InitConfig("testing")

	// Initialise config for all packages that interact
	// with either log or cache files
	util.SetConfig(config)
	worker.SetConfig(config)
	logging.SetConfig(config)
	graphing.SetConfig(config)

	apiKeys = getAPIKeysForTesting()

	os.RemoveAll(config.CacheFolderLocation)
	// Create test directories for logs and data
	os.Mkdir(config.CacheFolderLocation, 0755)
	os.Mkdir(config.LogsFolderLocation, 0755)
	os.Mkdir(config.FinishedGraphsLocation, 0755)

	code := m.Run()

	os.RemoveAll(config.CacheFolderLocation)
	os.RemoveAll(config.LogsFolderLocation)
	os.RemoveAll(config.FinishedGraphsLocation)

	os.Exit(code)
}

func TestGetUsernameFromCacheFile(t *testing.T) {
	cntr := util.Controller{}
	apiKeys := getAPIKeysForTesting()
	jobs := make(chan worker.JobsStruct, 100)

	targetSteamID := "76561198130544932"
	expectedUsername := "nestororan100"

	os.Setenv("CURRTARGET", targetSteamID)

	_, err := worker.GetFriends(cntr, targetSteamID, apiKeys[0], 1, jobs)
	if err != nil {
		t.Error(err)
	}
	usernameOfTargetUser, err := worker.GetUsernameFromCacheFile(cntr, targetSteamID)
	assert.Nil(t, err)
	assert.Equal(t, expectedUsername, usernameOfTargetUser)
}

func TestEndToEndCrawlingAndGraphingFunctionalityWithOneUser(t *testing.T) {
	cntr := util.Controller{}

	targetSteamID := "76561198130544932"
	mockUrlMap := make(map[string]string)
	// mockUrlMap[targetSteamID] = "outputGraphFor76561198130544932"

	crawlerConfig := worker.CrawlerConfig{
		Level:    2,
		StatMode: false,
		TestKeys: false,
		Workers:  5,
		APIKeys:  getAPIKeysForTesting(),
	}

	err := worker.CrawlOneUser(targetSteamID, mockUrlMap, cntr, crawlerConfig)

	assert.Nil(t, err)
}
