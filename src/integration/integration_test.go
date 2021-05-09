// +build integration

package integration

import (
	"os"
	"testing"

	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/logging"
	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
	"github.com/stretchr/testify/assert"
)

var (
	config configuration.Info
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
	// os.Setenv("testing", "")
	// path, err := os.Getwd()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// os.Setenv("BWD", fmt.Sprintf("%s/../", path))

	// os.RemoveAll(fmt.Sprintf("%s../testData", os.Getenv("BWD")))

	config := configuration.InitConfig("testing")

	util.SetConfig(configuration.InitConfig("testing"))
	worker.SetConfig(configuration.InitConfig("testing"))
	logging.SetConfig(configuration.InitConfig("testing"))

	os.RemoveAll(config.CacheFolderLocation)
	// Create test directories for logs and data
	os.Mkdir(config.CacheFolderLocation, 0755)
	os.Mkdir(config.LogsFolderLocation, 0755)

	code := m.Run()

	os.RemoveAll(config.CacheFolderLocation)
	os.RemoveAll(config.LogsFolderLocation)

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
	// if err != nil {
	// 	if !test.shouldFail {
	// 		t.Errorf("didn't fail to get username for %s", test.steamID)
	// 	}
	// } else if test.shouldFail {
	// 	t.Error("caught misbehaving testcase",
	// 		"SteamID:", test.steamID,
	// 	)
	// }

	// for _, elem := range tests {
	// 	_, err := worker.GetUsernameFromCacheFile(elem.steamID)
	// 	if err != nil {
	// 		if !elem.shouldFail {
	// 			t.Errorf("didn't fail to get username for %s", elem.steamID)
	// 		}
	// 	} else if elem.shouldFail {
	// 		t.Error("caught misbehaving testcase",
	// 			"SteamID:", elem.steamID,
	// 		)
	// 	}
	// }
}
