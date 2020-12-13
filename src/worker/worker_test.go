package worker

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/steamFriendsGraphing/util"
)

type testInput struct {
	steamID    string
	APIKey     string
	shouldFail bool
}

func TestMain(m *testing.M) {
	os.Setenv("testing", "")
	os.RemoveAll("../testData")
	os.Mkdir("../testData", 0755)
	os.Mkdir("../testLogs", 0755)

	code := m.Run()

	os.RemoveAll("../testData")
	os.RemoveAll("../testLogs")
	os.Exit(code)
}

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

func TestInitWorkerConfig(t *testing.T) {
	_, err := InitWorkerConfig(4, 15)
	if err != nil {
		t.Error(err)
	}

	_, err = InitWorkerConfig(-2, 20)
	if err == nil {
		t.Errorf("failed to catch invalid levelCap of -2")
	}

	_, err = InitWorkerConfig(2, 625)
	if err == nil {
		t.Errorf("failed to catch invalid worker amount of of 625")
	}

	_, err = InitWorkerConfig(2, 0)
	if err == nil {
		t.Errorf("failed to catch invalid worker amount of 0")
	}
}

func TestGetFriends(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	jobs := make(chan JobsStruct, 100)

	var tests = []testInput{
		{"76561198282036055", apiKeys[rand.Intn(len(apiKeys))], false},
		{"76561198282036055", apiKeys[rand.Intn(len(apiKeys))], false},
		{"76561198081485934", "invalid key", true},
		{"7656119807862962", apiKeys[rand.Intn(len(apiKeys))], true},
		{"76561198271948679", apiKeys[rand.Intn(len(apiKeys))], false},
		{"7656119796028793", apiKeys[rand.Intn(len(apiKeys))], true},
		{"76561198144084014", apiKeys[rand.Intn(len(apiKeys))], false},
		{"11111111111111111", apiKeys[rand.Intn(len(apiKeys))], true},
		{"gibberish", apiKeys[rand.Intn(len(apiKeys))], true},
	}

	for _, testCase := range tests {
		os.Setenv("CURRTARGET", testCase.steamID)
		_, err := os.Create(fmt.Sprintf("../testLogs/%s.txt", testCase.steamID))

		if err != nil {
			log.Fatal(fmt.Sprintf("eee %s", err))
		}
		_, err = GetFriends(testCase.steamID, testCase.APIKey, 1, jobs)
		if err != nil {
			if !testCase.shouldFail {
				t.Error("Error:", err,
					"SteamID:", testCase.steamID,
				)
			}
		} else if testCase.shouldFail {
			t.Error("caught misbehaving testcase",
				"APIKEY:", testCase.APIKey,
				"SteamID:", testCase.steamID,
			)
		}
	}

}

func TestGetUsernameFromCacheFile(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	jobs := make(chan JobsStruct, 100)

	_, err := GetFriends("76561198096639661", apiKeys[0], 1, jobs)
	if err != nil {
		t.Error(err)
	}

	var tests = []testInput{
		{"76561198306145504", "", true},
		{"76561198096639661", "", false},
	}
	for _, elem := range tests {
		_, err := GetUsernameFromCacheFile(elem.steamID)
		if err != nil {
			if !elem.shouldFail {
				t.Errorf("didn't fail to get username for %s", elem.steamID)
			}
		} else if elem.shouldFail {
			t.Error("caught misbehaving testcase",
				"SteamID:", elem.steamID,
			)
		}
	}
}

func TestGetCache(t *testing.T) {
	_, err := GetCache("76561198282036055e")
	if err == nil {
		t.Error("invalid cache get did not throw an error")
	}

	// Will fail if not run as part of the whole test suite
	// as this cache file will not have been written
	_, err = GetCache("76561198245030292")
	if err == nil {
		t.Error("Got invalid cache for user 76561198245030292")
	}
}

func TestExampleInvocation(t *testing.T) {
	apiKeys := getAPIKeysForTesting()
	fmt.Printf("==================== Example Invocation ====================\n")
	ControlFunc(apiKeys, "76561198130544932", 2, 2)
	fmt.Printf("\n")
	ControlFunc(apiKeys, "76561198130544932", 1, 1)
	fmt.Printf("\n")
	ControlFunc(apiKeys, "76561198130544932", 1, 1)
	fmt.Printf("============================================================\n")
}

func TestConfigInit(t *testing.T) {
	APIKeys := getAPIKeysForTesting()
	// Regular invocation
	fmt.Printf("\n\n")
	testConfig := CrawlerConfig{
		Level:    1,
		StatMode: false,
		TestKeys: false,
		Workers:  1,
		APIKeys:  APIKeys,
	}
	InitCrawling(testConfig, "76561198282036055")
	fmt.Printf("\n")
	// StatMode invocation
	testConfig2 := CrawlerConfig{
		Level:    1,
		StatMode: true,
		TestKeys: false,
		Workers:  1,
		APIKeys:  APIKeys,
	}
	InitCrawling(testConfig2, "76561198144084014")
	fmt.Printf("\n")
	// testKeys invocation
	testConfig3 := CrawlerConfig{
		Level:    1,
		StatMode: false,
		TestKeys: true,
		Workers:  1,
		APIKeys:  APIKeys,
	}
	InitCrawling(testConfig3, "76561198144084014")
	fmt.Printf("\n\n")
}
