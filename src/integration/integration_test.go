// +build integration

package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/steamFriendsGraphing/configuration"
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
	configuration.InitAndSetConfig("testing", false)
	apiKeys = getAPIKeysForTesting()

	// Create test directories for logs and data
	os.Mkdir(configuration.AppConfig.CacheFolderLocation, 0755)
	os.Mkdir(configuration.AppConfig.LogsFolderLocation, 0755)
	os.Mkdir(configuration.AppConfig.FinishedGraphsLocation, 0755)

	code := m.Run()

	os.RemoveAll(configuration.AppConfig.CacheFolderLocation)
	os.RemoveAll(configuration.AppConfig.LogsFolderLocation)
	os.RemoveAll(configuration.AppConfig.FinishedGraphsLocation)

	os.Exit(code)
}

func TestGetUsernameFromCacheFile(t *testing.T) {
	cntr := util.Controller{}
	apiKeys := getAPIKeysForTesting()
	jobs := make(chan worker.JobsStruct, 100)

	targetSteamID := "76561198130544932"
	expectedUsername := "nestororan100"

	firstJob := worker.JobsStruct{
		OriginalTargetUserSteamID: targetSteamID,
		CurrentTargetSteamID:      targetSteamID,
		Level:                     1,
		APIKey:                    apiKeys[0],
	}
	fmt.Printf("the money: %s\n", configuration.AppConfig.CacheFolderLocation)
	_, err := worker.GetFriends(cntr, firstJob, 1, jobs)
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

func TestEndToEndCrawlingAndGraphingFunctionalityWithTwoUsers(t *testing.T) {
	cntr := util.Controller{}

	firstTargetSteamID := "76561198130544932"
	secondTargetSteamID := "76561198305082260"
	mockUrlMap := make(map[string]string)
	// mockUrlMap[targetSteamID] = "outputGraphFor76561198130544932"

	crawlerConfig := worker.CrawlerConfig{
		Level:    2,
		StatMode: false,
		TestKeys: false,
		Workers:  5,
		APIKeys:  getAPIKeysForTesting(),
	}

	err := worker.CrawlTwoUsers(firstTargetSteamID, secondTargetSteamID, mockUrlMap, cntr, crawlerConfig)

	assert.Nil(t, err)
}
