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

	config := worker.CrawlerConfig{
		Level:    *level,
		StatMode: *statMode,
		TestKeys: *testKeys,
		Workers:  *workers,
		SteamID:  os.Args[len(os.Args)-1],
		APIKeys:  apiKeys,
	}

	if len(os.Args) > 1 {
		cfg, err := worker.InitCrawling(config)
		if err != nil {
			log.Fatal(err)
		}
		graphing.InitGraphing(cfg.Level, cfg.Workers, cfg.SteamID)
	} else {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
	}
}
