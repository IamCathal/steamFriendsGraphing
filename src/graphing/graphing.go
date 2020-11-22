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

type GraphData struct {
	SteamID      string
	Nodes        []charts.GraphNode
	Links        []charts.GraphLink
	EchartsGraph *charts.Graph

	ApplyDijkstra bool
	UsersMap      map[int]string
	DijkstraGraph *dijkstra.Graph
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

// func MergeDijkstraNodes(startUserGraph, endUserGraph *dijkstra.Graph) *dijkstra.Graph {

// }

func (gData *GraphData) GetDijkstraPath(startUserID, endUserID string) []string {
	fmt.Printf("GETTING DIJKSTRA\n")
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
	fmt.Println(gData.UsersMap)
	fmt.Printf("%d -> %d\n", firstUser, secondUser)
	best, err := gData.DijkstraGraph.Shortest(firstUser, secondUser)
	bestPathUsernames := make([]string, 0)
	if err != nil {
		fmt.Println("Couldn't find a path")
	} else {
		fmt.Println("Shortest distance ", best.Distance, " following path ")

		for _, id := range best.Path {
			fmt.Printf("%s -> ", gData.UsersMap[id])
			bestPathUsernames = append(bestPathUsernames, gData.UsersMap[id])
		}
		fmt.Println("")
	}

	return bestPathUsernames
}

func MergeDijkstraGraphs(startUserGraph, endUserGraph *dijkstra.Graph, startUsersMap, endUsersMap map[int]string) (*dijkstra.Graph, map[int]string) {
	allUsersMap := mergeUsersMaps(startUsersMap, endUsersMap)
	fmt.Println("start users map::::::::")
	for key, val := range startUsersMap {
		fmt.Printf("[%d : %s]\n", key, val)
	}
	fmt.Println("")
	fmt.Println("end users map::::::::")
	for key, val := range endUsersMap {
		fmt.Printf("[%d : %s]\n", key, val)
	}
	fmt.Println("")
	fmt.Println("All users map::::::::")
	for key, val := range allUsersMap {
		fmt.Printf("[%d : %s]\n", key, val)
	}
	fmt.Println("")
	existingNodesInt := make(map[int]bool)
	allGraph := dijkstra.NewGraph()
	fmt.Printf("\n\n\n")
	for _, indivNode := range startUserGraph.Verticies {
		fmt.Printf("[1]Does node %d exist yet? ", indivNode.ID)
		if exists := NodeExistsInt(indivNode.ID, existingNodesInt); !exists {
			fmt.Printf("yes\n")
			allGraph.AddVertex(indivNode.ID)
			existingNodesInt[indivNode.ID] = true
			// fmt.Printf("Added %d: %+v\n", indivNode.ID, indivNode.Arcs)
		}
		for ID, _ := range indivNode.Arcs {
			// fmt.Printf("Added from %d: %d\n", indivNode.ID, ID)
			// fmt.Printf("Add %d -> %d AND %d -> %d\n", indivNode.ID, ID, ID, indivNode.ID)
			if exist := NodeExistsInt(ID, existingNodesInt); !exist {
				allGraph.AddVertex(ID)
			}

			allGraph.AddArc(indivNode.ID, ID, 1)
			allGraph.AddArc(ID, indivNode.ID, 1)
			// fmt.Println(allGraph.Verticies)
		}
		// fmt.Printf("After pass: %+v\n\n", startUserGraph.Verticies)
	}

	for _, indivNode := range endUserGraph.Verticies {
		if indivNode.ID != 0 {
			fmt.Printf("[2]Does node %d exist yet? ", indivNode.ID)
			convertedIDUsername := endUsersMap[indivNode.ID]
			convertedID, ok := GetKeyFromValue(allUsersMap, convertedIDUsername)
			if !ok {
				if convertedID == 0 {
					log.Fatal("BAD THIS SHOULD NEVER HAPPEN")
				}
			}
			fmt.Printf("\n============== %d is actually [%d] %s\n", indivNode.ID, convertedID, convertedIDUsername)
			if exists := NodeExistsInt(convertedID, existingNodesInt); !exists {
				allGraph.AddVertex(convertedID)
				existingNodesInt[convertedID] = true
				fmt.Printf("Added %d: %+v\n", convertedID, indivNode.Arcs)
			}
			for ID, _ := range indivNode.Arcs {
				// fmt.Printf("Added from %d: %d\n", indivNode.ID, ID)
				arcConvertedIDUsername := endUsersMap[ID]
				arcConvertedID, ok := GetKeyFromValue(allUsersMap, arcConvertedIDUsername)
				fmt.Printf("[[[[[[ %d is actually %d [%s]\n", ID, arcConvertedID, arcConvertedIDUsername)
				if !ok {
					if convertedID == 0 {
						log.Fatal("BAD THIS SHOULD NEVER HAPPEN")
					}
				}
				// fmt.Printf("Add %d -> %d AND %d -> %d\n", convertedID, arcConvertedID, arcConvertedID, convertedID)
				if exist := NodeExistsInt(arcConvertedID, existingNodesInt); !exist {
					allGraph.AddVertex(arcConvertedID)
				}

				allGraph.AddArc(convertedID, arcConvertedID, 1)
				allGraph.AddArc(arcConvertedID, convertedID, 1)
				// fmt.Println(allGraph.Verticies)
			}
		}
		// fmt.Printf("After pass: %+v\n\n", startUserGraph.Verticies)
	}

	// fmt.Printf("All graph after first passs:\n")
	// for _, indivNode := range allGraph.Verticies {
	// 	fmt.Printf("--- %d:\n", indivNode.ID)
	// 	for ID, _ := range indivNode.Arcs {
	// 		fmt.Printf("[%d] arc %d -> %d\n", len(indivNode.Arcs), indivNode.ID, ID)
	// 	}
	// }
	// fmt.Printf("\n")

	// for _, indivNode := range endUserDijkstraGraph.Verticies {
	// 	allGraph.AddVertex(indivNode.ID)
	// 	for ID, _ := range indivNode.Arcs {
	// 		fmt.Printf("Adding %d -> %d\n", indivNode.ID, ID)
	// 		allGraph.AddArc(indivNode.ID, ID, 1)
	// 	}
	// }

	// fmt.Printf("All graph after second passs:\n")
	// for _, indivNode := range allGraph.Verticies {
	// 	fmt.Printf("%d: %d\n", indivNode.ID, indivNode.Arcs)
	// }

	// for _, indivNode := range endUserDijkstraGraph.Verticies {
	// 	allGraph.AddVertex(indivNode.ID)
	// }

	// for _, indivNode := range endUserDijkstraGraph.Verticies {
	// 	for ID, _ := range indivNode.Arcs {
	// 		// fmt.Printf("Adding %d -> %d\n", indivNode.ID, ID)
	// 		allGraph.AddArc(indivNode.ID, ID, 1)
	// 	}
	// }
	fmt.Println("START VERTICIEDS")
	for _, indivNode := range startUserGraph.Verticies {
		fmt.Printf("%d: %d\n", indivNode.ID, indivNode.Arcs)
	}

	fmt.Println("END VERTICIES")
	for _, indivNode := range endUserGraph.Verticies {
		fmt.Printf("%d: %d\n", indivNode.ID, indivNode.Arcs)
	}
	fmt.Println("ALL GRAPH")
	for _, indivNode := range allGraph.Verticies {
		fmt.Printf("%d: %d\n", indivNode.ID, indivNode.Arcs)
	}

	return allGraph, allUsersMap
}

func (gData *GraphData) Render() {
	fmt.Println("Rendering")

	gData.EchartsGraph.SetGlobalOptions(charts.TitleOpts{Title: "Yop the ladeens 示例图"},
		charts.InitOpts{Width: "1800px", Height: "1080px"})

	gData.EchartsGraph.Add("graph", gData.Nodes, gData.Links,
		charts.GraphOpts{Layout: "force", Roam: true, Force: charts.GraphForce{Repulsion: 34, Gravity: 0.16}, FocusNodeAdjacency: true},
		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"}},
		charts.LineStyleOpts{Width: 1, Color: "#b5b5b5"},
	)

	err := CreateFinishedGraphFolder()
	CheckErr(err)
	file, err := os.Create("../finishedGraphs/eee.html")
	CheckErr(err)

	gData.EchartsGraph.Render(file)
	fmt.Println("Wrote to file")
}

func InitGraphing(level, workers int, steamID string) *GraphData {
	fmt.Printf("=============================================\n")
	fmt.Printf("                GRAPHING\n\n")
	username, err := GetUsernameFromCacheFile(steamID)
	CheckErr(err)

	return CrawlCachedFriends(level, workers, steamID, username)
}
