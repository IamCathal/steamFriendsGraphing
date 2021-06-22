package worker

import (
	"fmt"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/graphing"
	"github.com/steamFriendsGraphing/util"
)

// CrawlOneUser crawls a single user and generates a graph the specified users friend network
func CrawlOneUser(steamID string, cntr util.ControllerInterface, config CrawlerConfig) error {
	finishedGraphLocation := ""

	if userHasBeenGraphedBefore := util.IsKeyInUrlMap(steamID); !userHasBeenGraphedBefore {
		GenerateURL(steamID)

		InitCrawling(cntr, config, steamID)
		gData, err := graphing.InitGraphing(config.Level, config.Workers, steamID)
		if err != nil {
			return err
		}

		finishedGraphLocation = fmt.Sprintf("%s/%s", configuration.AppConfig.FinishedGraphsLocation, configuration.AppConfig.UrlMap[steamID])
		err = gData.Render(finishedGraphLocation)
		if err != nil {
			return err
		}
	}

	finishedGraphLocation = fmt.Sprintf("%s/%s", configuration.AppConfig.FinishedGraphsLocation, configuration.AppConfig.UrlMap[steamID])
	fmt.Printf("Saved as %s.html\n", finishedGraphLocation)
	return nil
}

// CrawlOneUser crawls two users and generates a unified graph of their friend networks if possible
func CrawlTwoUsers(steamID1, steamID2 string, urlMapping map[string]string, cntr util.ControllerInterface, config CrawlerConfig) error {
	steamIDsIdentifier, err := getSteamIDsIdentifier([]string{steamID1, steamID2}, urlMapping)
	if err != nil {
		return err
	}
	finishedGraphLocation := ""

	if usersHaveBeenGraphedBefore := util.IsKeyInUrlMap(steamIDsIdentifier); !usersHaveBeenGraphedBefore {
		GenerateURL(steamIDsIdentifier)
		finishedGraphLocation = fmt.Sprintf("%s/%s", configuration.AppConfig.FinishedGraphsLocation, urlMapping[steamIDsIdentifier])

		InitCrawling(cntr, config, steamID1)
		InitCrawling(cntr, config, steamID2)

		StartUserGraphData, err := graphing.InitGraphing(config.Level, config.Workers, steamID1)
		if err != nil {
			return err
		}
		EndUserGraphData, err := graphing.InitGraphing(config.Level, config.Workers, steamID2)
		if err != nil {
			return err
		}

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
		graphData.Render(finishedGraphLocation)
	}

	finishedGraphLocation = fmt.Sprintf("%s/%s", configuration.AppConfig.FinishedGraphsLocation, urlMapping[steamIDsIdentifier])
	fmt.Printf("Saved as %s.html\n", finishedGraphLocation)
	return nil
}
