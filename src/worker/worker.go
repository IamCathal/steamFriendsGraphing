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
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/logging"
	"github.com/steamFriendsGraphing/util"
)

// JobsStruct is the strcuture used for placing jobs onto the
// worker queue
type JobsStruct struct {
	Level                     int
	OriginalTargetUserSteamID string
	CurrentTargetSteamID      string
	APIKey                    string
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

// InitWorkerConfig initialises the worker based on the level and worker amount given
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
	// fmt.Printf("======================================\n")
	// fmt.Printf("       Crawler configuration\n")
	// fmt.Printf("Level:\t\t%d\n", levelCap)
	// fmt.Printf("Worker amount:\t%d\n", workerAmount)
	// if levelCap <= 2 {
	// 	fmt.Printf("Channel len:\t700\n")
	// } else {
	// 	fmt.Printf("Channel len:\t%d\n", int(math.Pow(90, float64(levelCap))))
	// }
	// fmt.Printf("======================================\n")
	return workConfig, nil
}

// InitCrawling initialises the crawling and then starts up the graph crawler
// that produces the HTML output
func InitCrawling(cntr util.ControllerInterface, cfg CrawlerConfig, steamID string) {
	if cfg.TestKeys {
		util.CheckAPIKeys(cntr, cfg.APIKeys)
	}

	ControlFunc(cntr, cfg.APIKeys, steamID, cfg.Level, cfg.Workers)
}

// Worker is the crawling worker queue implementation. It takes in users off the jobs queue, processes
//  them and places and then places user's friends onto the results queue
func Worker(cntr util.ControllerInterface, jobs <-chan JobsStruct, results chan<- JobsStruct, cfg *WorkerConfig, activeJobs *int64) {
	for {
		cfg.JobsMutex.Lock()
		job := <-jobs
		cfg.JobsMutex.Unlock()

		// Temporary fix, sometimes level 0s get put onto jobs queue
		if job.Level != 0 {
			friendsObj, err := GetFriends(cntr, job, cfg.LevelCap, jobs)
			util.CheckErr(err)

			numFriends := len(friendsObj.FriendsList.Friends)

			// For each friend we'll add them to the jobs queue if
			// their level is within our range
			for i := 0; i < numFriends; i++ {
				indivFriends := JobsStruct{
					OriginalTargetUserSteamID: job.OriginalTargetUserSteamID,
					Level:                     job.Level + 1,
					CurrentTargetSteamID:      friendsObj.FriendsList.Friends[i].Steamid,
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
func GetFriends(cntr util.ControllerInterface, job JobsStruct, level int, jobs <-chan JobsStruct) (util.FriendsStruct, error) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)

	exists, err := CacheFileExists(cntr, job.CurrentTargetSteamID)
	if exists {
		if !configuration.AppConfig.IgnoreCache {
			friendsObj, err := GetCache(cntr, job.CurrentTargetSteamID)
			if err != nil {
				return util.FriendsStruct{}, err
			}
			LogCall(cntr, "GET", job, friendsObj.Username, "200", util.Green, startTime)
			return friendsObj, nil
		}
	}
	if err != nil {
		return util.FriendsStruct{}, err
	}

	// Check to see if the steamID is in the valid format now to save time
	if valid := util.IsValidFormatSteamID(job.CurrentTargetSteamID); !valid {
		LogCall(cntr, "GET", job, "Invalid SteamID", "400", util.Red, startTime)
		return util.FriendsStruct{}, util.MakeErr(fmt.Errorf("invalid steamID: %s, apikey: %s", job.CurrentTargetSteamID, job.APIKey))
	}
	friendsObj, err := cntr.CallGetFriendsListAPI(url.QueryEscape(job.CurrentTargetSteamID), url.QueryEscape(job.APIKey))
	if err != nil {
		LogCall(cntr, "GET", job, friendsObj.Username, "400", util.Red, startTime)
		return util.FriendsStruct{}, err
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
		userStatsObj, err := cntr.CallPlayerSummaryAPI(steamIDsList, job.APIKey)
		if err != nil {
			var temp util.FriendsStruct
			return temp, util.MakeErr(err)
		}
		// Do NOT process users with communityvisibilitystate = 1, this means they're private and will fuck up the system

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

			userStatsObj, err := cntr.CallPlayerSummaryAPI(steamIDsList, job.APIKey)
			if err != nil {
				var temp util.FriendsStruct
				return temp, util.MakeErr(err)
			}

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

	username, err := util.GetUsername(cntr, job.APIKey, job.CurrentTargetSteamID)
	if err != nil {
		return friendsObj, util.MakeErr(err)
	}
	friendsObj.Username = username
	WriteToFile(cntr, job.APIKey, job.CurrentTargetSteamID, friendsObj)
	// log the request along the round trip delay
	LogCall(cntr, fmt.Sprintf("GET [%d][%d]", level, len(jobs)), job, friendsObj.Username, "200", util.Green, startTime)
	return friendsObj, nil
}

// ControlFunc is the parent function of Worker. It adds the target user to the jobs queue and then processes the
// results queue until all users below the target level have been crawled
func ControlFunc(cntr util.ControllerInterface, apiKeys []string, steamID string, levelCap, workerAmount int) {
	workConfig, err := InitWorkerConfig(levelCap, workerAmount)
	if err != nil {
		log.Fatal(err)
	}
	logMsg := ""

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
		go Worker(cntr, jobs, results, workConfig, &activeJobs)
	}

	tempStruct := JobsStruct{
		OriginalTargetUserSteamID: steamID,
		Level:                     1,
		CurrentTargetSteamID:      steamID,
		APIKey:                    apiKeys[0],
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
				OriginalTargetUserSteamID: steamID,
				Level:                     result.Level,
				CurrentTargetSteamID:      result.CurrentTargetSteamID,
				APIKey:                    apiKeys[i%len(apiKeys)],
			}
			workConfig.Wg.Add(1)
			jobs <- newJob
			i++

		}
	}

	workConfig.Wg.Wait()
	logMsg += "\n=============== Done ================\n"
	logMsg += fmt.Sprintf("Total friends: %d\nCrawled friends: %d\n", totalFriends, reachableFriends)
	logMsg += fmt.Sprintf("Friends per level: %+v\n=====================================\n", friendsPerLevel)
	close(jobs)
	close(results)

	logFileName := fmt.Sprintf("%s/%s.txt", configuration.AppConfig.LogsFolderLocation, configuration.AppConfig.UrlMap[steamID])
	logging.SpecialLog(cntr, logFileName, logMsg)
}

// Divmod divides a friendslist into stacks of 100 and the remainder
func Divmod(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

// LogCall logs a call to the API with various stats on the request
func LogCall(cntr util.ControllerInterface, method string, job JobsStruct, username, status, statusColor string, startTime int64) {
	endTime := time.Now().UnixNano() / int64(time.Millisecond)
	delay := strconv.FormatInt((endTime - startTime), 10)

	logMsg := fmt.Sprintf("%s [%s] %s %s%s%s %vms\n", method, job.CurrentTargetSteamID, username,
		statusColor, status, "\033[0m", delay)
	logging.SpecialLog(cntr, configuration.AppConfig.UrlMap[job.OriginalTargetUserSteamID], logMsg)
	// fmt.Printf("%s", logMsg)
}

// WriteToFile writes a user's friendlist to a file for later processing
func WriteToFile(cntr util.ControllerInterface, apiKey, steamID string, friends util.FriendsStruct) error {
	cacheFolder := configuration.AppConfig.CacheFolderLocation
	if cacheFolder == "" {
		return util.MakeErr(errors.New("configuration.AppConfig.CacheFolderLocation was not initialised before attempting to write to file"))
	}

	existing, err := CacheFileExists(cntr, steamID)
	if !existing {
		file, err := cntr.CreateFile(fmt.Sprintf("%s/%s.gz", cacheFolder, steamID))
		if err != nil {
			return util.MakeErr(err)
		}

		jsonObj, err := json.Marshal(friends)
		if err != nil {
			return util.MakeErr(err)
		}
		err = cntr.WriteGzip(file, string(jsonObj))
		if err != nil {
			return util.MakeErr(err)
		}

		err = file.Close()
		if err != nil {
			return util.MakeErr(err)
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// GetCache gets a user's cached records if it exists
func GetCache(cntr util.ControllerInterface, steamID string) (util.FriendsStruct, error) {
	var temp util.FriendsStruct
	cacheFolder := configuration.AppConfig.CacheFolderLocation

	exists, err := CacheFileExists(cntr, steamID)
	if exists {
		file, err := cntr.Open(fmt.Sprintf("%s/%s.gz", cacheFolder, steamID))
		if err != nil {
			return temp, util.MakeErr(err)
		}
		gz, err := gzip.NewReader(file)
		if err != nil {
			return temp, util.MakeErr(err)
		}

		s, err := ioutil.ReadAll(gz)
		if err != nil {
			return temp, util.MakeErr(err)
		}

		err = json.Unmarshal([]byte(s), &temp)
		if err != nil {
			return temp, util.MakeErr(err)
		}
		err = file.Close()
		if err != nil {
			return temp, util.MakeErr(err)
		}
		err = gz.Close()
		if err != nil {
			return temp, util.MakeErr(err)
		}

		return temp, nil
	}
	if err != nil {
		return temp, err
	}

	return temp, util.MakeErr(fmt.Errorf("cache file %s/%s.gz does not exist", configuration.AppConfig.CacheFolderLocation, steamID))
}

// GetUsernameFromCacheFile gets the username for a given cache file
// e.g 76561198063271448 -> moose
func GetUsernameFromCacheFile(cntr util.ControllerInterface, steamID string) (string, error) {
	var temp util.FriendsStruct
	cacheFolder := configuration.AppConfig.CacheFolderLocation
	if cacheFolder == "" {
		return "", util.MakeErr(errors.New("configuration.AppConfig.CacheFolderLocation was not initialised before attempting to write to file"))
	}

	exists, err := CacheFileExists(cntr, steamID)
	if exists {
		file, err := os.Open(fmt.Sprintf("%s/%s.gz", cacheFolder, steamID))
		if err != nil {
			return "", util.MakeErr(err)
		}
		gz, err := gzip.NewReader(file)
		if err != nil {
			return "", util.MakeErr(err)
		}
		scanner := bufio.NewScanner(gz)

		res := ""
		for scanner.Scan() {
			res += scanner.Text()
		}
		err = json.Unmarshal([]byte(res), &temp)
		if err != nil {
			return "", util.MakeErr(err)
		}
		err = file.Close()
		if err != nil {
			return "", util.MakeErr(err)
		}
		err = gz.Close()
		if err != nil {
			return "", util.MakeErr(err)
		}

		return temp.Username, nil
	}
	if err != nil {
		return "", err
	}

	return "", util.MakeErr(fmt.Errorf("cache file %s.gz does not exist", steamID))
}

// CacheFileExists checks whether a given cached file exists
func CacheFileExists(cntr util.ControllerInterface, steamID string) (bool, error) {
	cacheFolder := configuration.AppConfig.CacheFolderLocation
	if cacheFolder == "" {
		return false, util.MakeErr(errors.New("configuration.AppConfig.CacheFolderLocation was not initialised before attempting to write to file"))
	}

	return cntr.FileExists(fmt.Sprintf("%s/%s.gz", cacheFolder, steamID)), nil
}
