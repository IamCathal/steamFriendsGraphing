package graphing

import (
	"os"
	"testing"

	"github.com/go-echarts/go-echarts/charts"
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
	os.Mkdir("../testData", 0755)

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

func TestMergeNodes(t *testing.T) {
	nodes1 := make([]charts.GraphNode, 0)
	nodes2 := make([]charts.GraphNode, 0)

	nodes1 = append(nodes1, charts.GraphNode{Name: "Cathal"})
	nodes1 = append(nodes1, charts.GraphNode{Name: "Joe"})
	nodes1 = append(nodes1, charts.GraphNode{Name: "Declan"})
	nodes1 = append(nodes1, charts.GraphNode{Name: "Michael"})

	nodes2 = append(nodes2, charts.GraphNode{Name: "Michael"})
	nodes2 = append(nodes2, charts.GraphNode{Name: "Declan"})
	nodes2 = append(nodes2, charts.GraphNode{Name: "Johnny"})
	nodes2 = append(nodes2, charts.GraphNode{Name: "Mairtin"})

	allNodes := MergeNodes(nodes1, nodes2)
	wantNodes := []string{"Cathal", "Joe", "Declan", "Michael", "Johnny", "Mairtin"}

	for i, name := range wantNodes {
		if allNodes[i].Name != name {
			t.Errorf("Node %d has %s instead of %s", i, allNodes[i].Name, name)
		}
	}
}
