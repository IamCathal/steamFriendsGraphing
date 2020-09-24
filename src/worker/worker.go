package worker

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/steamFriendsGraphing/util"
)

// JobsStruct is the strcuture used for placing jobs onto the
// worker queue
type JobsStruct struct {
	Level   int
	SteamID string
	APIKey  string
}

// WorkerConfig holds most of the configuration needed
// to start the worker (apart from jobs and result channels)
type WorkerConfig struct {
	JobsMutex       *sync.Mutex
	ResMutex        *sync.Mutex
	ActiveJobsMutex *sync.Mutex
	Wg              *sync.WaitGroup
	Jobs            *chan JobsStruct
	Results         *chan JobsStruct
	LevelCap        int
	WorkerAmount    int
}

// CrawlerConfig holdes all of the configuration needed to
// start the crawler
type CrawlerConfig struct {
	Level    int
	StatMode bool
	TestKeys bool
	Workers  int
	APIKeys  []string
}

// InitWorkerConfig initialised the worker based on the levle and worker amount given
// and returns a WorkerConfig with all fields set
func InitWorkerConfig(levelCap, workerAmount int) (*WorkerConfig, error) {

	if levelCap < 1 || levelCap > 4 {
		temp := &WorkerConfig{}
		return temp, fmt.Errorf("invalid level %d given. levelCap must be in range 1-4 (inclusive)", levelCap)
	}

	if workerAmount < 1 || workerAmount > 60 {
		temp := &WorkerConfig{}
		return temp, fmt.Errorf("invalid worker amount %d given. worker amount must be in range 1-60 (inclusive)", levelCap)
	}

	var wg sync.WaitGroup
	var resMutex sync.Mutex
	var activeJobsMutex sync.Mutex
	var jobMutex sync.Mutex

	workConfig := &WorkerConfig{
		JobsMutex:       &jobMutex,
		ResMutex:        &resMutex,
		ActiveJobsMutex: &activeJobsMutex,
		Wg:              &wg,
		LevelCap:        levelCap,
		WorkerAmount:    workerAmount,
	}
	fmt.Printf("======================================\n")
	fmt.Printf("       Crawler configuration\n")
	fmt.Printf("Level:\t\t%d\n", levelCap)
	fmt.Printf("Worker amount:\t%d\n", workerAmount)
	if levelCap <= 2 {
		fmt.Printf("Channel len:\t700\n")
	} else {
		fmt.Printf("Channel len:\t%d\n", int(math.Pow(90, float64(levelCap))))
	}
	fmt.Printf("======================================\n")
	return workConfig, nil
}

// InitCrawling initialises the crawling and then starts up the graph crawler
// that produces the HTML output
func InitCrawling(cfg CrawlerConfig, steamID string) (CrawlerConfig, error) {
	temp := CrawlerConfig{Level: -1}
	if cfg.TestKeys == true {
		util.CheckAPIKeys(cfg.APIKeys)
		return temp, nil
	}

	if cfg.StatMode {
		resMap, err := util.GetUserDetails(cfg.APIKeys[0], steamID)
		if err != nil {
			return temp, err
		}
		for k, v := range resMap {
			fmt.Printf("%13s: %s\n", k, v)
		}
		return temp, nil
	}

	ControlFunc(cfg.APIKeys, steamID, cfg.Level, cfg.Workers)
	return cfg, nil
}

// Worker is the crawling worker queue implementation. It takes in users off the jobs queue, processes
//  them and places and then places user's friends onto the results queue
func Worker(jobs <-chan JobsStruct, results chan<- JobsStruct, cfg *WorkerConfig, activeJobs *int64) {
	for {
		cfg.JobsMutex.Lock()
		job := <-jobs
		cfg.JobsMutex.Unlock()

		// Temporary fix, sometimes level 0s get put onto jobs queue
		if job.Level != 0 {
			friendsObj, err := GetFriends(job.SteamID, job.APIKey, cfg.LevelCap, jobs)
			util.CheckErr(err)

			numFriends := len(friendsObj.FriendsList.Friends)

			// For each friend we'll add them to the jobs queue if
			// their level is within our range
			for i := 0; i < numFriends; i++ {
				indivFriends := JobsStruct{
					Level:   job.Level + 1,
					SteamID: friendsObj.FriendsList.Friends[i].Steamid,
				}

				// If their level is within range, we'll scrape them in the future
				// and therefore we up the counter of activeJobs
				if indivFriends.Level <= cfg.LevelCap {
					atomic.AddInt64(activeJobs, 1)
				}

				cfg.ResMutex.Lock()
				results <- indivFriends
				cfg.ResMutex.Unlock()
			}

			cfg.ActiveJobsMutex.Lock()
			atomic.AddInt64(activeJobs, -1)
			cfg.ActiveJobsMutex.Unlock()

			cfg.Wg.Done()
		}
	}
}

// GetFriends returns the list of friends for a given user and caches results if requested
func GetFriends(steamID, apiKey string, level int, jobs <-chan JobsStruct) (util.FriendsStruct, error) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)

	// If the cache exists and the env var to disable serving from cache is not set
	if exists := CacheFileExists(steamID); exists {
		if exists := IsEnvVarSet("disablereadcache"); !exists {
			friendsObj, err := GetCache(steamID)
			if err != nil {
				return friendsObj, err
			}
			LogCall("GET", steamID, friendsObj.Username, "200", util.Green, startTime)
			return friendsObj, nil
		}
	}

	// Check to see if the steamID is in the valid format now to save time
	if valid := util.IsValidFormatSteamID(steamID); !valid {
		LogCall("GET", steamID, "Invalid SteamID", "400", util.Red, startTime)
		var temp util.FriendsStruct
		return temp, fmt.Errorf("invalid steamID %s, apikey %s\n", steamID, apiKey)
	}
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s&relationship=friend", url.QueryEscape(apiKey), url.QueryEscape(steamID))
	res, err := http.Get(targetURL)
	util.CheckErr(err)
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	util.CheckErr(err)

	var friendsObj util.FriendsStruct
	json.Unmarshal(body, &friendsObj)

	// If the HTTP response has error messages in it handle them accordingly
	if valid := util.IsValidSteamID(string(body)); !valid {
		LogCall("GET", steamID, friendsObj.Username, "400", util.Red, startTime)
		var temp util.FriendsStruct
		return temp, errors.New("Invalid steamID given")
	}

	if valid := util.IsValidAPIKey(string(body)); !valid {
		LogCall("GET", steamID, "Invalid API key", "403", util.Red, startTime)
		var temp util.FriendsStruct
		return temp, fmt.Errorf("invalid api key: %s", apiKey)
	}

	// Gathers usernames from steamIDs
	steamIDsList := ""
	friendsListLen := len(friendsObj.FriendsList.Friends)

	// Only 100 steamIDs can be given per call, hence we must
	// divide the friends list into lists of 100 or less
	callCount, remainder := Divmod(friendsListLen, 100)

	// less than 100 friends, only 1 call is needed
	if callCount < 1 {
		for ind, val := range friendsObj.FriendsList.Friends {
			if ind < friendsListLen-1 {
				steamIDsList += val.Steamid + ","
			} else {
				steamIDsList += val.Steamid
			}
		}
		steamIDsList += ""

		targetURL = fmt.Sprintf("http://api.steampowered.com/I SteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamIDsList)
		res, err = http.Get(targetURL)
		util.CheckErr(err)
		body, err = ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		util.CheckErr(err)

		var userStatsObj util.UserStatsStruct
		json.Unmarshal(body, &userStatsObj)

		// Order of received friends is random,
		// must assign them using map[steamID]username
		friendsMap := make(map[string]string)
		for _, user := range userStatsObj.Response.Players {
			friendsMap[user.Steamid] = user.Personaname
		}

		for i := 0; i < len(friendsObj.FriendsList.Friends); i++ {
			friendsObj.FriendsList.Friends[i].Username = friendsMap[friendsObj.FriendsList.Friends[i].Steamid]
		}

	} else {
		// More than 100 friends, subsequent calls are needed
		for i := 0; i <= callCount; i++ {
			//each batch of 100

			steamIDsList = ""

			if i < callCount {
				// a full call of 100 friends
				for k := 0; k < 100; k++ {
					steamIDsList += friendsObj.FriendsList.Friends[k+(i*100)].Steamid + ","
				}
			} else {
				// a batch of the remainder (less than 100)
				for k := 0; k < remainder; k++ {
					steamIDsList += friendsObj.FriendsList.Friends[k+(i*100)].Steamid + ","
				}
			}

			targetURL = fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamIDsList)
			res, err = http.Get(targetURL)
			util.CheckErr(err)
			body, err = ioutil.ReadAll(res.Body)
			defer res.Body.Close()
			util.CheckErr(err)

			var userStatsObj util.UserStatsStruct
			json.Unmarshal(body, &userStatsObj)

			// Order of received friends is random,
			// must assign them using map[steamID]username
			friendsMap := make(map[string]string)
			for _, user := range userStatsObj.Response.Players {
				friendsMap[user.Steamid] = user.Personaname
			}

			if i < callCount {
				for k := 0; k < 100; k++ {
					// find the entry in the friendsObj struct and set the username field
					friendsObj.FriendsList.Friends[k+(i*100)].Username = friendsMap[friendsObj.FriendsList.Friends[k+(i*100)].Steamid]
				}
			} else {
				for k := 0; k < remainder; k++ {
					friendsObj.FriendsList.Friends[k+(i*100)].Username = friendsMap[friendsObj.FriendsList.Friends[k+(i*100)].Steamid]
				}
			}

		}

	}

	username, err := util.GetUsername(apiKey, steamID)
	util.CheckErr(err)
	friendsObj.Username = username

	WriteToFile(apiKey, steamID, friendsObj)

	// log the request along the round trip delay
	LogCall(fmt.Sprintf("GET [%d][%d]", level, len(jobs)), steamID, friendsObj.Username, "200", util.Green, startTime)
	return friendsObj, nil
}

// ControlFunc is the parent function of Worker. It adds the target user to the jobs queue and then processes the
// results queue until all users below the target level have been crawled
func ControlFunc(apiKeys []string, steamID string, levelCap, workerAmount int) {
	workConfig, err := InitWorkerConfig(levelCap, workerAmount)
	if err != nil {
		log.Fatal(err)
	}

	// After level 3 the amount of friends gets CRAZY
	// Therefore some rapid scaling is needed
	// Level 2: 8100 buffer length
	// Level 3: 729000 buffer length
	// Level 4: 6.561e+07 buffer length (This is not feasible to crawl)
	chanLen := 0
	if levelCap <= 2 {
		chanLen = 700
	} else {
		chanLen = int(math.Pow(90, float64(levelCap)))
	}
	jobs := make(chan JobsStruct, chanLen)
	results := make(chan JobsStruct, chanLen)

	var activeJobs int64 = 0
	friendsPerLevel := make(map[int]int)

	for i := 0; i < workConfig.WorkerAmount; i++ {
		go Worker(jobs, results, workConfig, &activeJobs)
	}

	tempStruct := JobsStruct{
		Level:   1,
		SteamID: steamID,
		APIKey:  apiKeys[0],
	}

	workConfig.Wg.Add(1)
	activeJobs++
	jobs <- tempStruct
	friendsPerLevel[1]++

	reachableFriends := 0
	totalFriends := 0
	i := 1

	for {
		if activeJobs == 0 {
			break
		}
		result := <-results
		totalFriends++
		friendsPerLevel[result.Level]++

		if result.Level <= levelCap {
			reachableFriends++

			newJob := JobsStruct{
				Level:   result.Level,
				SteamID: result.SteamID,
				APIKey:  apiKeys[i%len(apiKeys)],
			}
			workConfig.Wg.Add(1)
			jobs <- newJob
			i++

		}
	}

	workConfig.Wg.Wait()
	fmt.Printf("\n=============== Done ================\n")
	fmt.Printf("Total friends: %d\nCrawled friends: %d\n", totalFriends, reachableFriends)
	fmt.Printf("Friends per level: %+v\n=====================================\n", friendsPerLevel)
	close(jobs)
	close(results)

}

// Divmod divides a friendslist into stacks of 100 and the remainder
func Divmod(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

// LogCall logs a call to the API with various stats on the request
func LogCall(method, steamID, username, status, statusColor string, startTime int64) {
	endTime := time.Now().UnixNano() / int64(time.Millisecond)
	delay := strconv.FormatInt((endTime - startTime), 10)

	fmt.Printf("%s [%s] %s %s%s%s %vms\n", method, steamID, username,
		statusColor, status, "\033[0m", delay)
}

// WriteToFile writes a user's friendlist to a file for later processing
func WriteToFile(apiKey, steamID string, friends util.FriendsStruct) {
	cacheFolder := "userData"
	if exists := util.IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}

	if existing := CacheFileExists(steamID); !existing {
		file, err := os.Create(fmt.Sprintf("../%s/%s.gz", cacheFolder, steamID))
		fmt.Printf("[Worker WriteToFile] Here is ../\n")
		files, err := ioutil.ReadDir("../")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			fmt.Println(f.Name())
		}
		fmt.Printf("\n\n")
		fmt.Printf("[Worker WriteToFile] Here is ../testData\n")
		files, err = ioutil.ReadDir("../testData")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			fmt.Println(f.Name())
		}
		fmt.Printf("\n\n")
		util.CheckErr(err)
		defer file.Close()

		jsonObj, err := json.Marshal(friends)
		util.CheckErr(err)

		w := gzip.NewWriter(file)
		w.Write([]byte(jsonObj))
		w.Close()
	}
}

// GetCache gets a user's cached records if it exists
func GetCache(steamID string) (util.FriendsStruct, error) {
	var temp util.FriendsStruct
	cacheFolder := "userData"
	if exists := util.IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}

	if exists := CacheFileExists(steamID); exists {
		file, err := os.Open(fmt.Sprintf("../%s/%s.gz", cacheFolder, steamID))
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}

		gz, _ := gzip.NewReader(file)
		defer gz.Close()

		scanner := bufio.NewScanner(gz)
		res := ""
		for scanner.Scan() {
			res += scanner.Text()
		}

		_ = json.Unmarshal([]byte(res), &temp)
		return temp, nil
	}

	return temp, fmt.Errorf("Cache file %s.gz does not exist", steamID)
}

// GetUsernameFromCacheFile gets the username for a given cache file
// e.g 76561198063271448 -> moose
func GetUsernameFromCacheFile(steamID string) (string, error) {
	var temp util.FriendsStruct
	cacheFolder := "userData"
	if exists := util.IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}

	if exists := CacheFileExists(steamID); exists {
		file, err := os.Open(fmt.Sprintf("../%s/%s.gz", cacheFolder, steamID))
		defer file.Close()
		if err != nil {
			return "", err
		}

		gz, _ := gzip.NewReader(file)
		defer gz.Close()

		scanner := bufio.NewScanner(gz)
		res := ""
		for scanner.Scan() {
			res += scanner.Text()
		}

		_ = json.Unmarshal([]byte(res), &temp)
		return temp.Username, nil
	}

	return "", fmt.Errorf("Cache file %s.gz does not exist", steamID)
}

// IsEnvVarSet does a simple check to see if an environment
// variable is set
func IsEnvVarSet(envvar string) bool {
	if _, exists := os.LookupEnv(envvar); exists {
		return true
	}
	return false
}

// CacheFileExists checks whether a given cached file exists
func CacheFileExists(steamID string) bool {
	cacheFolder := "userData"
	if exists := IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}

	_, err := os.Stat(fmt.Sprintf("../%s/%s.gz", cacheFolder, steamID))
	if os.IsNotExist(err) {
		return false
	}
	return true
}
