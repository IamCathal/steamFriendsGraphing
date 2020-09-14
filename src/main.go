package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"
)

// IndivFriend holds profile information for
// a single user
type IndivFriend struct {
	Steamid string `json:"steamid"`
	// Steamid is the default steamid64 value for a user
	Relationship string `json:"relationship"`
	// Relationship is always friend in this case
	FriendSince int64 `json:"friend_since,omitempty"`
	// FriendSince is the unix timestamp of when the friend request was accepted
	Username string `json:"username"`
	// Username is steam public username
}

// FriendsStruct is messy but it holds the array of all friends for a given user
type FriendsStruct struct {
	Username    string `json:"username"`
	FriendsList struct {
		Friends []IndivFriend `json:"friends"`
	} `json:"friendslist"`
}

// Player holds all account information for a given user. This is the
// response from the getPlayerSummaries endpoint
type Player struct {
	Steamid                  string `json:"steamid"`
	Communityvisibilitystate int    `json:"communityvisibilitystate"`
	Profilestate             int    `json:"profilestate"`
	Personaname              string `json:"personaname"`
	Profileurl               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	Avatarmedium             string `json:"avatarmedium"`
	Avatarfull               string `json:"avatarfull"`
	Personastate             int    `json:"personastate"`
	Realname                 string `json:"realname,omitempty"`
	Primaryclanid            string `json:"primaryclanid,omitempty"`
	Timecreated              int    `json:"timecreated,omitempty"`
	Personastateflags        int    `json:"personastateflags,omitempty"`
	Loccountrycode           string `json:"loccountrycode,omitempty"`
	Commentpermission        int    `json:"commentpermission,omitempty"`
}

// UserStatsStruct is the JSON response of the
// userstats lookup to get usernames from steamIDs
type UserStatsStruct struct {
	Response struct {
		Players []struct {
			Player
		} `json:"players"`
	} `json:"response"`
}

// GetFriends returns the list of friends for a given user and caches results if requested
func GetFriends(steamID, apiKey string, level int, jobs <-chan jobsStruct) (FriendsStruct, error) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)

	// If the cache exists and the env var to disable serving from cache is not set
	if exists := CacheFileExist(steamID); exists {
		if exists := IsEnvVarSet("disablereadcache"); !exists {
			friendsObj, err := GetCache(steamID)
			if err != nil {
				return friendsObj, err
			}
			LogCall("GET", steamID, friendsObj.Username, "200", green, startTime)
			return friendsObj, nil
		}
	}

	// Check to see if the steamID is in the valid format now to save time
	if valid := IsValidFormatSteamID(steamID); !valid {
		LogCall("GET", steamID, "Invalid SteamID", "400", red, startTime)
		var temp FriendsStruct
		return temp, fmt.Errorf("invalid steamID %s, apikey %s\n", steamID, apiKey)
	}

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s&relationship=friend", url.QueryEscape(apiKey), url.QueryEscape(steamID))
	res, err := http.Get(targetURL)
	CheckErr(err)
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	CheckErr(err)

	var friendsObj FriendsStruct
	json.Unmarshal(body, &friendsObj)

	// If the HTTP response has error messages in it handle them accordingly
	if valid := IsValidSteamID(string(body)); !valid {
		LogCall("GET", steamID, friendsObj.Username, "400", red, startTime)
		var temp FriendsStruct
		return temp, errors.New("Invalid steamID given")
	}

	if valid := IsValidAPIKey(string(body)); !valid {
		LogCall("GET", steamID, "Invalid API key", "403", red, startTime)
		var temp FriendsStruct
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

		targetURL = fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamIDsList)
		res, err = http.Get(targetURL)
		CheckErr(err)
		body, err = ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		CheckErr(err)

		var userStatsObj UserStatsStruct
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
			CheckErr(err)
			body, err = ioutil.ReadAll(res.Body)
			defer res.Body.Close()
			CheckErr(err)

			var userStatsObj UserStatsStruct
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

	username, err := GetUsername(apiKey, steamID)
	CheckErr(err)
	friendsObj.Username = username

	WriteToFile(apiKey, steamID, friendsObj)

	// log the request along the round trip delay
	LogCall(fmt.Sprintf("GET [%d][%d]", level, len(jobs)), steamID, friendsObj.Username, "200", green, startTime)
	return friendsObj, nil
}

func newControlFunc(apiKeys []string, steamID string, levelCap, workerAmount int) {
	workConfig, err := InitWorkerConfig(levelCap, workerAmount)
	if err != nil {
		log.Fatal(err)
	}

	// After level 3 the amount of friends gets CRAZY
	// Therefore some rapid scaling is needed
	// Level 2: 8100 buffer length
	// Level 3: 729000 buffer length
	// Level 4: 6.561e+07 buffer length (This is not feasable to crawl)
	chanLen := 0
	if levelCap <= 2 {
		chanLen = 700
	} else {
		chanLen = int(math.Pow(90, float64(levelCap)))
	}
	jobs := make(chan jobsStruct, chanLen)
	results := make(chan jobsStruct, chanLen)

	var activeJobs int64 = 0
	friendsPerLevel := make(map[int]int)

	for i := 0; i < workConfig.workerAmount; i++ {
		go Worker(jobs, results, workConfig, &activeJobs)
	}

	tempStruct := jobsStruct{
		level:   1,
		steamID: steamID,
		APIKey:  apiKeys[0],
	}

	workConfig.wg.Add(1)
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
		friendsPerLevel[result.level]++

		if result.level <= levelCap {
			reachableFriends++

			newJob := jobsStruct{
				level:   result.level,
				steamID: result.steamID,
				APIKey:  apiKeys[i%len(apiKeys)],
			}
			workConfig.wg.Add(1)
			jobs <- newJob
			i++

		}
	}

	workConfig.wg.Wait()
	fmt.Printf("\n=============== Done ================\n")
	fmt.Printf("Total friends: %d\nCrawled friends: %d\n", totalFriends, reachableFriends)
	fmt.Printf("Friends per level: %+v\n=====================================\n", friendsPerLevel)
	close(jobs)
	close(results)

}

func main() {

	level := flag.Int("level", 2, "Level of friends you want to crawl. 2 is your friends, 3 is mutual friends etc")
	statMode := flag.Bool("stat", false, "Simple lookup of a target user.")
	testKeys := flag.Bool("testkeys", false, "Test if all keys in APIKEYS.txt are valid")
	workerAmount := flag.Int("workeramount", 2, "Amount of workers that are crawling")
	flag.Parse()

	apiKeys, err := GetAPIKeys()
	CheckErr(err)

	if *testKeys == true {
		CheckAPIKeys(apiKeys)
		return
	}

	if *statMode {
		err := PrintUserDetails(apiKeys[0], os.Args[len(os.Args)-1])
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if len(os.Args) > 1 {
		if *level == 1 {
			*statMode = true
		}
		// Last argument should be the steamID
		newControlFunc(apiKeys, os.Args[len(os.Args)-1], *level, *workerAmount)
	} else {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
	}
}
