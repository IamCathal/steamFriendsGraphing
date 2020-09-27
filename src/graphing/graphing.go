package graphing

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RyanCarrier/dijkstra"
	"github.com/go-echarts/go-echarts/charts"
)

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

// graphWorker is the graphing worker queue implementation. It's quite similar to
// the crawling worker in the worker module but this is purely for graphing
func graphWorker(id int, jobs <-chan infoStruct, results chan<- infoStruct, wConfig *workerConfig, gConfig *graphConfig, wg *sync.WaitGroup, activeJobs *int64, levelCap int) {
	for {
		wConfig.jobMutex.Lock()
		job := <-jobs
		wConfig.jobMutex.Unlock()
		rand.Seed(time.Now().UTC().UnixNano())

		if job.level != 0 {
			friendsObj, err := GetCache(job.steamID)
			CheckErr(err)

			friendCount := len(friendsObj.FriendsList.Friends)

			// Iterate through the user's friendlist and add them onto the
			// results channel for future processing. Important to link
			// the current user and this friend so a link can be made later
			for i := 0; i < friendCount; i++ {
				tempStruct := infoStruct{
					level:    job.level + 1,
					from:     job.username,
					steamID:  friendsObj.FriendsList.Friends[i].Steamid,
					username: friendsObj.FriendsList.Friends[i].Username,
				}

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

func CrawlCachedFriends(level, workers int, steamID, username string) {

	makePath := false

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

	for i := 0; i < workers*2; i++ {
		go graphWorker(i, jobs, results, &wConfig, &gConfig, &wg, &activeJobs, levelCap)
	}

	tempStruct := infoStruct{
		level:    1,
		steamID:  steamID,
		username: username,
		from:     username,
	}

	usersCount := 1
	// Users is used to map a users position in the stack of calls to their username
	// This is used as the dijkstra implementation only sorts based on ints so each
	// user must be assigned this as a key and then coverted back later into usernames
	users := make(map[int]string)
	users[usersCount] = tempStruct.username

	dijkstraGraph := dijkstra.NewGraph()
	dijkstraGraph.AddVertex(usersCount)

	usersCount++

	gConfig.existingNodes[tempStruct.username] = true
	specColor := charts.ItemStyleOpts{Color: "#000000"}
	gConfig.nodes = append(gConfig.nodes, charts.GraphNode{Name: tempStruct.username, ItemStyle: specColor})

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
		totalFriends++
		friendsPerLevel[result.level]++

		if result.level <= levelCap {
			reachableFriends++

			if exists := NodeExists(result.username, existingNodes); !exists {
				gConfig.existingNodes[result.username] = true
				gConfig.nodes = append(gConfig.nodes, charts.GraphNode{Name: result.username})

				users[usersCount] = result.username
				dijkstraGraph.AddVertex(usersCount)
				usersCount++
			}
			gConfig.links = append(gConfig.links, charts.GraphLink{Source: result.from, Target: result.username})

			fromNum, ok := GetKeyFromValue(users, result.from)
			if !ok {
				log.Fatal("BAD THIS SHOULD NEVER HAPPEN")
			}
			dijkstraGraph.AddArc(fromNum, usersCount-1, 1)

			fmt.Printf("[%d] %s[%s] -> %s[%s]\n", result.level, result.from, result.steamID, result.username, result.steamID)
			newJob := infoStruct{
				level:    result.level,
				steamID:  result.steamID,
				from:     result.from,
				username: result.username,
			}
			wg.Add(1)
			if newJob.from == "" {
				log.Fatalf("Empty job caught: %+v", newJob)
			}
			jobs <- newJob
		}

	}

	wg.Wait()
	fmt.Printf("\n============== Done ==============\n")
	close(jobs)
	close(results)

	firstUser, ok := GetKeyFromValue(users, "Omac995")
	if !ok {
		fmt.Printf("User Omac995 has not been crawled\n")
	}

	secondUser, ok := GetKeyFromValue(users, "lozefase")
	if !ok {
		fmt.Printf("User lozefase has not been crawled\n")
	}

	if makePath == true {
		best, err := dijkstraGraph.Shortest(firstUser, secondUser)
		if err != nil {
			fmt.Println("Couldn't find a path")
		} else {
			fmt.Println("Shortest distance ", best.Distance, " following path ")

			for _, id := range best.Path {
				fmt.Printf("%s -> ", users[id])
			}
			fmt.Println("")
		}
	}

	graph.SetGlobalOptions(charts.TitleOpts{Title: "Yop the ladeens 示例图"},
		charts.InitOpts{Width: "1800px", Height: "1080px"})

	graph.Add("graph", gConfig.nodes, gConfig.links,
		charts.GraphOpts{Layout: "force", Roam: true, Force: charts.GraphForce{Repulsion: 34, Gravity: 0.16}, FocusNodeAdjacency: true},
		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"}},
		charts.LineStyleOpts{Width: 1, Color: "#b5b5b5"},
	)

	err := CreateFinishedGraphFolder()
	CheckErr(err)
	file, err := os.Create(fmt.Sprintf("../finishedGraphs/%s.html", steamID))
	CheckErr(err)

	graph.Render(file)
}

func InitGraphing(level, workers int, steamID string) {
	fmt.Printf("=============================================\n")
	fmt.Printf("                GRAPHING\n\n")
	username, err := GetUsernameFromCacheFile(steamID)
	CheckErr(err)

	CrawlCachedFriends(level, workers, steamID, username)
}
