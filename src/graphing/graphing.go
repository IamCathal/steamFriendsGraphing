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
	dijkstra "github.com/iamcathal/dijkstra2"
	"github.com/steamFriendsGraphing/configuration"
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

// GraphData holds all of the data points needed to a friend network
// graph using go-echarts
type GraphData struct {
	SteamID      string
	Nodes        []charts.GraphNode
	Links        []charts.GraphLink
	EchartsGraph *charts.Graph

	ApplyDijkstra bool
	UsersMap      map[int]string
	DijkstraGraph *dijkstra.Graph
}

var (
	appConfig configuration.Info
)

func SetConfig(config configuration.Info) {
	appConfig = config
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
			// results channel for future processing.
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

// CrawlCachedFriends builds the graph structure from cached users
func CrawlCachedFriends(level, workers int, steamID, username string) *GraphData {
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
	// Give the original user a black colored node to stand out
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
			dijkstraGraph.AddArc(usersCount-1, fromNum, 1)

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

	gData := &GraphData{
		SteamID:      steamID,
		Nodes:        gConfig.nodes,
		Links:        gConfig.links,
		EchartsGraph: graph,

		UsersMap:      users,
		DijkstraGraph: dijkstraGraph,
	}
	return gData
}

func mergeUsersMaps(startUsersMap, endUsersMap map[int]string) map[int]string {
	allUsersMap := make(map[int]string)
	for key, val := range startUsersMap {
		allUsersMap[key] = val
	}
	for _, val := range endUsersMap {
		if _, exists := GetKeyFromValue(allUsersMap, val); !exists {
			allUsersMap[len(allUsersMap)+1] = val
		}
	}
	return allUsersMap
}

// GetDijkstraPath gets the actual shortest path (if possible) between two given users
func (gData *GraphData) GetDijkstraPath(startUserID, endUserID string) ([]string, bool) {
	// Convert username to it's associated ID in the dijkstra graph.
	// This conversion must be carried out because the dijkstra implementation
	// only works based off of the ID field which is an int
	firstUsername, err := GetUsernameFromCacheFile(startUserID)
	CheckErr(err)
	firstUser, ok := GetKeyFromValue(gData.UsersMap, firstUsername)
	if !ok {
		fmt.Printf("User %s has not been crawled\n", firstUsername)
	}
	secondUsername, err := GetUsernameFromCacheFile(endUserID)
	CheckErr(err)
	secondUser, ok := GetKeyFromValue(gData.UsersMap, secondUsername)
	if !ok {
		fmt.Printf("User %s has not been crawled\n", secondUsername)
	}

	best, err := gData.DijkstraGraph.Shortest(firstUser, secondUser)
	bestPathUsernames := make([]string, 0)
	if err != nil {
		return []string{}, false
	} else {
		fmt.Println("Shortest distance ", best.Distance, " following path ")

		for _, id := range best.Path {
			bestPathUsernames = append(bestPathUsernames, gData.UsersMap[id])
		}
	}
	return bestPathUsernames, true
}

// MergeDijkstraGraph merges the dijkstra graphs of two users into one logical graph. This
// enables the application to find the shortest past if any users are present on both graphs
func MergeDijkstraGraphs(startUserGraph, endUserGraph *dijkstra.Graph, startUsersMap, endUsersMap map[int]string) (*dijkstra.Graph, map[int]string) {
	allUsersMap := mergeUsersMaps(startUsersMap, endUsersMap)

	existingNodesInt := make(map[int]bool)
	allGraph := dijkstra.NewGraph()
	for _, indivNode := range startUserGraph.Verticies {
		if exists := NodeExistsInt(indivNode.ID, existingNodesInt); !exists {
			allGraph.AddVertex(indivNode.ID)
			existingNodesInt[indivNode.ID] = true
		}
		for ID, _ := range indivNode.Arcs {
			if exist := NodeExistsInt(ID, existingNodesInt); !exist {
				allGraph.AddVertex(ID)
			}
			allGraph.AddArc(indivNode.ID, ID, 1)
			allGraph.AddArc(ID, indivNode.ID, 1)
		}
	}

	for _, indivNode := range endUserGraph.Verticies {
		if indivNode.ID != 0 {
			convertedIDUsername := endUsersMap[indivNode.ID]
			convertedID, ok := GetKeyFromValue(allUsersMap, convertedIDUsername)
			if !ok {
				if convertedID == 0 {
					log.Fatal("BAD THIS SHOULD NEVER HAPPEN")
				}
			}
			if exists := NodeExistsInt(convertedID, existingNodesInt); !exists {
				allGraph.AddVertex(convertedID)
				existingNodesInt[convertedID] = true
			}
			for ID, _ := range indivNode.Arcs {
				arcConvertedIDUsername := endUsersMap[ID]
				arcConvertedID, ok := GetKeyFromValue(allUsersMap, arcConvertedIDUsername)
				if !ok && convertedID == 0 {
					log.Fatal("BAD THIS SHOULD NEVER HAPPEN")
				}
				if exist := NodeExistsInt(arcConvertedID, existingNodesInt); !exist {
					allGraph.AddVertex(arcConvertedID)
				}

				allGraph.AddArc(convertedID, arcConvertedID, 1)
				allGraph.AddArc(arcConvertedID, convertedID, 1)
			}
		}
	}

	return allGraph, allUsersMap
}

// Render generates the HTML graph output
func (gData *GraphData) Render(fileName string) {
	gData.EchartsGraph.SetGlobalOptions(charts.TitleOpts{Title: "Yop the ladeens 薄煎饼"},
		charts.InitOpts{Width: "1800px", Height: "1080px"})

	gData.EchartsGraph.Add("graph", gData.Nodes, gData.Links,
		charts.GraphOpts{Layout: "force", Roam: true, Force: charts.GraphForce{Repulsion: 34, Gravity: 0.16}, FocusNodeAdjacency: true},
		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"}},
		charts.LineStyleOpts{Width: 1, Color: "#b5b5b5"},
	)
	err := CreateFinishedGraphFolder()
	CheckErr(err)
	file, err := os.Create(fmt.Sprintf("%s.html", fileName))
	CheckErr(err)
	gData.EchartsGraph.Render(file)
}

// InitGraphing kicks off the graphing process
func InitGraphing(level, workers int, steamID string) (*GraphData, error) {
	fmt.Printf("=============================================\n")
	fmt.Printf("                GRAPHING\n\n")
	username, err := GetUsernameFromCacheFile(steamID)

	return CrawlCachedFriends(level, workers, steamID, username), err
}
