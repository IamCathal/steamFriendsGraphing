package worker

import (
	"fmt"
	"os"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/steamFriendsGraphing/graphing"
	"github.com/steamFriendsGraphing/util"
)

func CrawlOneUser(steamID string, urlMapping map[string]string, cntr util.ControllerInterface, config CrawlerConfig) {
	finishedGraphLocation := fmt.Sprintf("%s/%s", appConfig.FinishedGraphsLocation, urlMapping[steamID])
	os.Setenv("CURRTARGET", steamID)

	if userHasBeenGraphedBefore := util.IfKeyNotInMap(steamID, urlMapping); !userHasBeenGraphedBefore {
		GenerateURL(steamID, urlMapping)

		InitCrawling(cntr, config, steamID)
		gData := graphing.InitGraphing(config.Level, config.Workers, steamID)
		gData.Render(finishedGraphLocation)
	}

	fmt.Printf("Saved as %s.html\n", finishedGraphLocation)
}

func CrawlTwoUsers(steamID1, steamID2 string, urlMapping map[string]string, cntr util.ControllerInterface, config CrawlerConfig) {
	identifier := fmt.Sprintf("%s%s", steamID1, steamID2)
	GenerateURL(fmt.Sprintf("%s%s", steamID1, steamID2), urlMapping)

	os.Setenv("CURRTARGET", steamID1)
	InitCrawling(cntr, config, steamID1)

	os.Setenv("CURRTARGET", steamID2)
	InitCrawling(cntr, config, steamID2)

	StartUserGraphData := graphing.InitGraphing(config.Level, config.Workers, steamID1)
	EndUserGraphData := graphing.InitGraphing(config.Level, config.Workers, steamID2)

	graph := charts.NewGraph()
	allNodes := graphing.MergeNodes(StartUserGraphData.Nodes, EndUserGraphData.Nodes)

	allDijkstraGraph, allUsersMap := graphing.MergeDijkstraGraphs(StartUserGraphData.DijkstraGraph, EndUserGraphData.DijkstraGraph, StartUserGraphData.UsersMap, EndUserGraphData.UsersMap)

	graphData := &graphing.GraphData{
		Nodes:        allNodes,
		Links:        append(StartUserGraphData.Links, EndUserGraphData.Links...),
		EchartsGraph: graph,

		ApplyDijkstra: true,
		UsersMap:      allUsersMap,
		DijkstraGraph: allDijkstraGraph,
	}
	newNodes := make([]charts.GraphNode, 0)
	bestPath, aPathExists := graphData.GetDijkstraPath(steamID1, steamID2)

	if aPathExists {
		fmt.Println("The route:")
		for _, username := range bestPath {
			fmt.Printf("%s -> ", username)
		}

		fmt.Printf("\n")
		foundNode := false

		if len(bestPath) != 0 {
			for _, username := range graphData.Nodes {
				for _, pathUsername := range bestPath {
					if username.Name == pathUsername {
						specColor := charts.ItemStyleOpts{Color: "#38413A"}
						newNodes = append(newNodes, charts.GraphNode{Name: pathUsername, ItemStyle: specColor})
						foundNode = true
						break
					}
				}
				if !foundNode {
					newNodes = append(newNodes, charts.GraphNode{Name: username.Name})
				}
				foundNode = false
			}
		}
	}

	graphData.Nodes = newNodes
	finishedGraphLocation := fmt.Sprintf("%s/%s", appConfig.FinishedGraphsLocation, urlMapping[identifier])
	graphData.Render(finishedGraphLocation)

	fmt.Printf("Saved as %s.html\n", finishedGraphLocation)
}
