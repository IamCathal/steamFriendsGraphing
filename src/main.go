package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/server"
	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
)

func main() {
	level := flag.Int("level", 2, "Level of friends you want to crawl. 1 is just one user, 2 is immediate friends, 3 is mutual friends etc")
	statMode := flag.Bool("stat", false, "Perform a simple lookup of one user to retrieve basic profile details ")
	testKeys := flag.Bool("testkeys", false, "Test if all keys in APIKEYS.txt are valid")
	workers := flag.Int("workers", 2, "Amount of workers used to crawl")
	httpserver := flag.Bool("httpserver", false, "Run the application as a HTTP server")
	ignorecache := flag.Bool("ignorecache", false, "Don't read from cache")
	flag.Parse()

	cntr := util.Controller{}
	configuration.InitAndSetConfig("normal", *ignorecache)

	if *httpserver {
		server.SetController(cntr)
		server.RunServer("8080")
		return
	}

	apiKeys, err := util.GetAPIKeys(cntr)
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

	if len(os.Args) < 1 {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
		return
	}

	if config.StatMode {
		for _, steamID := range steamIDs {
			userDetails, err := util.GetUserDetails(cntr, config.APIKeys[0], steamID)
			util.CheckErr(err)

			fmt.Printf("SteamID:\t%s\n", userDetails.Steamid)
			fmt.Printf("Username:\t%s\n", userDetails.Personaname)
			fmt.Printf("TimeCreated:\t%s\n", time.Unix(int64(userDetails.Timecreated), 0))
			fmt.Printf("ProfileURL:\t%s\n", userDetails.Profileurl)
			fmt.Printf("AvatarURL:\t%s\n", userDetails.Avatarfull)
			fmt.Printf("\n")
		}
		return
	}

	// If two steamIDs are given and they are the same, treat this as a
	// single user search
	if len(steamIDs) == 2 && steamIDs[0] == steamIDs[1] {
		steamIDs = []string{steamIDs[0]}
	}

	switch len(steamIDs) {
	case 1:
		err := worker.CrawlOneUser(steamIDs[0], cntr, config)
		if err != nil {
			panic(err)
		}
	case 2:
		err := worker.CrawlTwoUsers(steamIDs[0], steamIDs[1], configuration.AppConfig.UrlMap, cntr, config)
		if err != nil {
			panic(err)
		}
	}

}
