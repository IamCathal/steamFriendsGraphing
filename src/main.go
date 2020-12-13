package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/steamFriendsGraphing/graphing"
	"github.com/steamFriendsGraphing/server"
	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
)

func main() {

	level := flag.Int("level", 2, "Level of friends you want to crawl. 2 is your friends, 3 is mutual friends etc")
	statMode := flag.Bool("stat", false, "Simple lookup of a target user.")
	testKeys := flag.Bool("testkeys", false, "Test if all keys in APIKEYS.txt are valid")
	workers := flag.Int("workers", 2, "Amount of workers that are crawling")
	httpserver := flag.Bool("httpserver", false, "Run the application as a HTTP server")
	flag.Parse()

	if *httpserver {
		server.RunServer("8080")
		return
	}
	apiKeys, err := util.GetAPIKeys()
	util.CheckErr(err)

	steamIDs, err := util.ExtractSteamIDs(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	config := worker.CrawlerConfig{
		Level:    *level,
		StatMode: *statMode,
		TestKeys: *testKeys,
		Workers:  *workers,
		APIKeys:  apiKeys,
	}

	urlMap := worker.LoadMappings()

	if len(os.Args) < 1 {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
		return
	}

	if config.StatMode {
		for _, steamID := range steamIDs {
			resMap, err := util.GetUserDetails(config.APIKeys[0], steamID)
			util.CheckErr(err)

			for k, v := range resMap {
				fmt.Printf("%13s: %s\n", k, v)
			}
			fmt.Printf("\n")
		}
		return
	}

	if len(steamIDs) == 1 {
		os.Setenv("CURRTARGET", steamIDs[0])
		worker.InitCrawling(config, steamIDs[0])
		worker.GenerateURL(steamIDs[0], urlMap)

		gData := graphing.InitGraphing(config.Level, config.Workers, steamIDs[0])
		gData.Render(urlMap[steamIDs[0]])
		return
	}

	if len(steamIDs) == 2 {
		identifier := fmt.Sprintf("%s%s", steamIDs[0], steamIDs[1])
		worker.GenerateURL(fmt.Sprintf("%s%s", steamIDs[0], steamIDs[1]), urlMap)

		os.Setenv("CURRTARGET", steamIDs[0])
		worker.InitCrawling(config, steamIDs[0])
		os.Setenv("CURRTARGET", steamIDs[1])
		worker.InitCrawling(config, steamIDs[1])

		StartUserGraphData := graphing.InitGraphing(config.Level, config.Workers, steamIDs[0])
		EndUserGraphData := graphing.InitGraphing(config.Level, config.Workers, steamIDs[1])

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
		bestPath, exists := graphData.GetDijkstraPath(steamIDs[0], steamIDs[1])
		if exists {
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
		graphData.Render(urlMap[identifier])
		return
	}
}
