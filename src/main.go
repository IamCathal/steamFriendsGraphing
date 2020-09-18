package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/steamFriendsGraphing/util"
	"github.com/steamFriendsGraphing/worker"
)

type config struct {
	level    int
	statMode bool
	testKeys bool
	workers  int
	steamID  string
	APIKeys  []string
}

func (cfg config) InitCrawling() {
	if cfg.testKeys == true {
		util.CheckAPIKeys(cfg.APIKeys)
		return
	}

	if cfg.statMode {
		err := util.PrintUserDetails(cfg.APIKeys[0], cfg.steamID)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Last argument should be the steamID
	worker.ControlFunc(cfg.APIKeys, cfg.steamID, cfg.level, cfg.workers)

}

func main() {

	level := flag.Int("level", 2, "Level of friends you want to crawl. 2 is your friends, 3 is mutual friends etc")
	statMode := flag.Bool("stat", false, "Simple lookup of a target user.")
	testKeys := flag.Bool("testkeys", false, "Test if all keys in APIKEYS.txt are valid")
	workers := flag.Int("workers", 2, "Amount of workers that are crawling")
	flag.Parse()

	apiKeys, err := util.GetAPIKeys()
	util.CheckErr(err)

	config := config{
		level:    *level,
		statMode: *statMode,
		testKeys: *testKeys,
		workers:  *workers,
		steamID:  os.Args[len(os.Args)-1],
		APIKeys:  apiKeys,
	}

	if len(os.Args) > 1 {
		config.InitCrawling()
	} else {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
	}
}
