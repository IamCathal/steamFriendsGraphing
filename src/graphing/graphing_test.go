// +build service

package graphing

import (
	"os"
	"testing"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/steamFriendsGraphing/configuration"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	appConfig = configuration.InitConfig("testing")
	SetConfig(appConfig)

	code := m.Run()

	os.Exit(code)
}

// func TestGraphing(t *testing.T) {
// 	APIKeys := getAPIKeysForTesting()
// 	files, err := ioutil.ReadDir("../")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("TESTGRAPHING THE FILES IN ABOVE DIR")
// 	for _, f := range files {
// 		fmt.Println(f.Name())
// 	}
// 	fmt.Printf("\n\n")
// 	// Must first fetch the data, otherwise there would
// 	// be no cached files to construct the graph with
// 	testConfig := worker.CrawlerConfig{
// 		Level:    2,
// 		StatMode: false,
// 		TestKeys: false,
// 		Workers:  1,
// 		APIKeys:  APIKeys,
// 	}
// 	worker.InitCrawling(testConfig, "76561198090461077")

// 	InitGraphing(2, 2, "76561198090461077")
// }

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

	actualNodes := MergeNodes(nodes1, nodes2)
	expectedNodes := []string{"Cathal", "Joe", "Declan", "Michael", "Johnny", "Mairtin"}

	actualNodeNames := []string{}
	for i, _ := range actualNodes {
		actualNodeNames = append(actualNodeNames, actualNodes[i].Name)
	}

	assert.Equal(t, expectedNodes, actualNodeNames)
}

func TestNodeExistsInt(t *testing.T) {
	targetID := 6
	nodeMap := make(map[int]bool, 0)

	nodeMap[targetID] = true

	if exists := NodeExistsInt(6, nodeMap); !exists {
		t.Errorf("Failed to find ID %d in %+v", targetID, nodeMap)
	}
	targetID = 74
	if exists := NodeExistsInt(targetID, nodeMap); exists {
		t.Errorf("Found non existant ID %d in %+v", targetID, nodeMap)
	}
}

func TestMergeUsersMaps(t *testing.T) {
	startUsersMap := make(map[int]string, 0)
	endUsersMap := make(map[int]string, 0)

	startUsersMap[0] = "Rob Pike"
	startUsersMap[1] = "Robert Griesemer"
	startUsersMap[2] = "Ken Thompson"
	endUsersMap[0] = "Guido van Rossum"
	endUsersMap[1] = "Rob Pike"
	endUsersMap[2] = "Yukihiro Matsumoto"
	actualUsersMap := mergeUsersMaps(startUsersMap, endUsersMap)

	expectedUsersMap := make(map[int]string)
	expectedUsersMap[0] = "Rob Pike"
	expectedUsersMap[1] = "Robert Griesemer"
	expectedUsersMap[2] = "Ken Thompson"
	expectedUsersMap[4] = "Guido van Rossum"
	expectedUsersMap[5] = "Yukihiro Matsumoto"

	allActualUsers := []string{}
	for _, name := range actualUsersMap {
		allActualUsers = append(allActualUsers, name)
	}
	allExpectedUsers := []string{}
	for _, name := range expectedUsersMap {
		allExpectedUsers = append(allExpectedUsers, name)
	}

	assert.ElementsMatch(t, allExpectedUsers, allActualUsers)
}

// func TestRender(t *testing.T) {
// 	graph := charts.NewGraph()
// 	nodes := make([]charts.GraphNode, 0)
// 	links := make([]charts.GraphLink, 0)

// 	nodes = append(nodes, charts.GraphNode{Name: "Rob Pike"})
// 	nodes = append(nodes, charts.GraphNode{Name: "Robert Griesemer"})
// 	nodes = append(nodes, charts.GraphNode{Name: "Ken Thompson"})

// 	links = append(links, charts.GraphLink{Source: "Rob Pike", Target: "Ken Thompson"})
// 	links = append(links, charts.GraphLink{Source: "Ken Thompson", Target: "Robert Griesemer"})

// 	graph.Add("graph", nodes, links,
// 		charts.GraphOpts{Layout: "force", Roam: true, Force: charts.GraphForce{Repulsion: 34, Gravity: 0.16}, FocusNodeAdjacency: true},
// 		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "white"}},
// 		charts.LineStyleOpts{Width: 1, Color: "#b5b5b5"},
// 	)

// 	graphData := &GraphData{
// 		Nodes:        nodes,
// 		Links:        links,
// 		EchartsGraph: graph,
// 	}

// 	graphData.Render(fmt.Sprintf("%s/../../finishedGraphs/testerGraph2", os.Getenv("BWD")))
// 	os.Remove(fmt.Sprintf("%s/../../finishedGraphs/testerGraph2.html", os.Getenv("BWD")))
// }
