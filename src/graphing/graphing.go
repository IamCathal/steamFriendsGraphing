package graphing

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-echarts/go-echarts/charts"
)

func graphWorker(id int, jobs <-chan infoStruct, results chan<- infoStruct, wConfig *workerConfig, gConfig *graphConfig, wg *sync.WaitGroup, activeJobs *int64, levelCap int) {
	for {
		wConfig.jobMutex.Lock()
		job := <-jobs
		wConfig.jobMutex.Unlock()
		rand.Seed(time.Now().UTC().UnixNano())

		if job.level != 0 {
			friendsObj, err := GetCache(job.steamID)
			if err != nil {
				log.Fatal(err)
			}

			friendCount := len(friendsObj.FriendsList.Friends)
			// fmt.Printf("Got [%d][%s][%s] - %d friends\n", job.level, friendsObj.Username, job.steamID, friendCount)

			// "Find" some friends and push them onto the jobs queue
			for i := 0; i < friendCount; i++ {
				tempStruct := infoStruct{
					level:    job.level + 1,
					from:     job.username,
					steamID:  friendsObj.FriendsList.Friends[i].Steamid,
					username: friendsObj.FriendsList.Friends[i].Username,
				}
				// fmt.Printf("\t%+v\n", tempStruct)
				// fmt.Printf("[%d] %s -> %s\n", tempStruct.level, friendsObj.Username, friendsObj.FriendsList.Friends[i].Username)

				if tempStruct.level <= levelCap {
					atomic.AddInt64(activeJobs, 1)
				}

				wConfig.resMutex.Lock()
				results <- tempStruct
				wConfig.resMutex.Unlock()

			}

			wConfig.activeJobsMutex.Lock()
			atomic.AddInt64(activeJobs, -1)
			wConfig.activeJobsMutex.Unlock()
			wg.Done()
		}
	}
}

func nodeExists(username string, nodeMap map[string]bool) bool {
	_, ok := nodeMap[username]
	if ok {
		return true
	}
	return false
}

type graphConfig struct {
	nodes         []charts.GraphNode
	links         []charts.GraphLink
	nodesMutex    *sync.Mutex
	linksMutex    *sync.Mutex
	existingNodes map[string]bool
}

type workerConfig struct {
	wg              *sync.WaitGroup
	jobMutex        *sync.Mutex
	resMutex        *sync.Mutex
	activeJobsMutex *sync.Mutex
}

func CrawlCachedFriends(level int, steamID, username string) {

	jobs := make(chan infoStruct, 500000)
	results := make(chan infoStruct, 500000)

	var wg sync.WaitGroup
	var jobMutex sync.Mutex
	var resMutex sync.Mutex
	var activeJobsMutex sync.Mutex

	wConfig := workerConfig{
		wg:              &wg,
		jobMutex:        &jobMutex,
		resMutex:        &resMutex,
		activeJobsMutex: &activeJobsMutex,
	}

	nodes := make([]charts.GraphNode, 0)
	links := make([]charts.GraphLink, 0)
	existingNodes := make(map[string]bool)
	var linkMutex sync.Mutex
	var nodeMutex sync.Mutex

	gConfig := graphConfig{
		links:         links,
		nodes:         nodes,
		nodesMutex:    &nodeMutex,
		linksMutex:    &linkMutex,
		existingNodes: existingNodes,
	}

	var activeJobs int64 = 0
	levelCap := level
	friendsPerLevel := make(map[int]int)

	for i := 0; i < 1; i++ {
		go graphWorker(i, jobs, results, &wConfig, &gConfig, &wg, &activeJobs, levelCap)
	}

	tempStruct := infoStruct{
		level:    1,
		steamID:  steamID,
		username: username,
		from:     username,
	}

	gConfig.existingNodes[tempStruct.username] = true
	gConfig.nodes = append(gConfig.nodes, charts.GraphNode{Name: tempStruct.username})

	wg.Add(1)
	activeJobs++
	jobs <- tempStruct
	friendsPerLevel[1]++

	reachableFriends := 0
	totalFriends := 0

	graph := charts.NewGraph()

	for {
		if activeJobs == 0 {
			break
		}
		result := <-results
		// fmt.Printf("\t\t%+v\n", result)
		// fmt.Printf("[%d] Got [SteamID %d][Level %d]\n", activeJobs, result.steamID, result.level)
		totalFriends++
		friendsPerLevel[result.level]++

		if result.level <= levelCap {
			reachableFriends++

			if exists := nodeExists(result.username, existingNodes); !exists {
				gConfig.existingNodes[result.username] = true
				gConfig.nodes = append(gConfig.nodes, charts.GraphNode{Name: result.username})
			}
			gConfig.links = append(gConfig.links, charts.GraphLink{Source: result.from, Target: result.username})
			fmt.Printf("[%d] %s[%s] -> %s[%s]\n", result.level, result.from, result.steamID, result.username, result.steamID)
			newJob := infoStruct{
				level:    result.level,
				steamID:  result.steamID,
				from:     result.from,
				username: result.username,
			}
			wg.Add(1)
			if newJob.from == "" {
				fmt.Printf("BADEEE %+v\n", newJob)
				panic(1)
			}
			// fmt.Printf("\t%+v\n", newJob)
			jobs <- newJob
		} else {
			// fmt.Printf("\t%+v\n", result)
		}

		// if result.level == levelCap {
		// 	if exists := nodeExists(result.username, existingNodes); !exists {
		// 		gConfig.nodes = append(gConfig.nodes, charts.GraphNode{Name: result.username})
		// 	} else {
		// 		gConfig.nodes = append(gConfig.nodes, charts.GraphNode{Name: result.username})

		// 	}
		// 	fmt.Printf("[%d] %s[%s] -> %s[%s]\n", result.level, result.from, result.steamID, result.username, result.steamID)

		// }

	}

	wg.Wait()

	// gConfig.links = append(gConfig.links, charts.GraphLink{Source: "disco biscuits", Target: "godanka"})

	fmt.Printf("%d --- %d \n", len(gConfig.nodes), len(gConfig.links))
	fmt.Printf("\n============== Done ==============\nTotal friends: %d\nCrawled friends: %d\n==================================\n", totalFriends, reachableFriends)
	fmt.Printf("%+v\n", friendsPerLevel)
	close(jobs)
	close(results)

	graph.SetGlobalOptions(charts.TitleOpts{Title: "Yop the ladeens 示例图"},
		charts.InitOpts{Width: "1800px", Height: "1080px"})
	graph.Add("graph", gConfig.nodes, gConfig.links,
		charts.GraphOpts{Layout: "force", Roam: true, Force: charts.GraphForce{Repulsion: 34, Gravity: 0.16}, FocusNodeAdjacency: true},
		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"}},
		charts.LineStyleOpts{Width: 1, Color: "#b5b5b5"},
	)

	file, err := os.Create("hello3.html")
	if err != nil {
		log.Println(err)
	}

	graph.Render(file)
}

func InitGraphing(level int, steamID string) {
	fmt.Printf("=============================================\n")
	fmt.Printf("                GRAPHING\n\n")
	username, err := GetUsernameFromCacheFile(steamID)
	if err != nil {
		log.Fatal(err)
	}

	CrawlCachedFriends(level, steamID, username)
}
