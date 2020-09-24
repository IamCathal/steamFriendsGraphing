package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	if len(os.Args) > 1 {
		for _, steamID := range steamIDs {
			cfg, err := worker.InitCrawling(config, steamID)
			if err != nil {
				log.Fatal(err)
			}
			// Level -1 is passed back when statMode is invoked
			if cfg.Level != -1 {
				graphing.InitGraphing(cfg.Level, cfg.Workers, steamID)
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
	}
}
