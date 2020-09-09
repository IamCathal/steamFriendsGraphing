package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// IndivFriend holds profile information for
// a single individual
type IndivFriend struct {
	Steamid string `json:"steamid"`
	// Steamid is the default steamid64 value for a user
	Relationship string `json:"relationship"`
	// Relationship is always friend in this case
	FriendSince int64 `json:"friend_since,omitempty"`
	// FriendSince, unix timestamp of when the friend request was accepted
	Username string `json:"username"`
	// Username is steam public username
}

// FriendsStruct messy but it holds the array of all friends
type FriendsStruct struct {
	Username    string `json:"username"`
	FriendsList struct {
		Friends []IndivFriend `json:"friends"`
	} `json:"friendslist"`
}

// Player holds all account information for a given player. This is the
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

// GetFriends Returns the list of friends for a given user
func GetFriends(steamID, apiKey string, waitG *sync.WaitGroup) (FriendsStruct, error) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	defer waitG.Done()

	// If the cache exists and the env var to disable serving from cache is set
	if exists := CacheFileExist(steamID); exists && os.Getenv("disablereadcache") != "" {
		friendsObj, err := GetCache(steamID)
		if err != nil {
			return friendsObj, err
		}
		go LogCall("GET", steamID, friendsObj.Username, "200", green, startTime)
		return friendsObj, nil
	}

	// Check to see if the steamID is in the valid format now to save time
	if valid := IsValidFormatSteamID(steamID); !valid {
		go LogCall("GET", steamID, "Invalid SteamID", "400", red, startTime)
		var temp FriendsStruct
		return temp, errors.New("Invalid steamID")
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
		go LogCall("GET", steamID, friendsObj.Username, "400", red, startTime)
		var temp FriendsStruct
		return temp, errors.New("Invalid steamID given")
	}

	if valid := IsValidAPIKey(string(body)); !valid {
		go LogCall("GET", steamID, "Invalid API key", "403", red, startTime)
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
					friendsObj.FriendsList.Friends[k+(i*100)].Username = friendsMap[friendsObj.FriendsList.Friends[k].Steamid]
				}
			}

		}

	}

	username, err := GetUsername(apiKey, steamID)
	CheckErr(err)
	friendsObj.Username = username

	WriteToFile(apiKey, steamID, friendsObj)

	// log the request along the round trip delay
	go LogCall("GET", steamID, friendsObj.Username, "200", green, startTime)
	return friendsObj, nil
}

// WriteToFile writes a user's friendlist to a file for later processing
func WriteToFile(apiKey, steamID string, friends FriendsStruct) {
	cacheFolder := "userData"
	if _, exists := os.LookupEnv("testing"); exists {
		cacheFolder = "testData"
	}

	if existing := CacheFileExist(steamID); !existing {
		fileLoc := fmt.Sprintf("../%s/%s.json", cacheFolder, steamID)
		file, err := os.Create(fileLoc)
		CheckErr(err)
		defer file.Close()

		jsonObj, err := json.Marshal(friends)
		CheckErr(err)

		_ = ioutil.WriteFile(fileLoc, jsonObj, 0644)
	}

}

// GetAPIKeys Retrieve the API key(s) to make requests with
// Keys must be stored in APIKEY(s).txt
func GetAPIKeys() ([]string, error) {
	file, err := os.Open("APIKEYS.txt")
	if err != nil {
		CheckErr(errors.New("No APIKEYS.txt file found"))
	}
	defer file.Close()

	apiKeys := make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		apiKeys = append(apiKeys, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("Error reading APIKEYS.txt")
	}

	if len(apiKeys) > 0 {
		return apiKeys, nil
	}
	// APIKeys.txt does exist but it is empty
	return nil, errors.New("No API key(s)")
}

func controlFunc(apiKeys []string, steamID string, statMode bool) {
	var waitG sync.WaitGroup
	waitG.Add(1)

	err := CreateUserDataFolder()
	CheckErr(err)

	friendsObj, err := GetFriends(steamID, apiKeys[0], &waitG)
	CheckErr(err)

	numFriends := len(friendsObj.FriendsList.Friends)

	if numFriends == 0 {
		fmt.Printf("User has previously been queried\n")
		return
	}

	fmt.Printf("Friends: %d\n", numFriends)
	if statMode {
		PrintUserDetails(apiKeys[0], steamID)
		return
	}

	for i, friend := range friendsObj.FriendsList.Friends {
		waitG.Add(1)
		go GetFriends(friend.Steamid, apiKeys[i%(len(apiKeys))], &waitG)
		// Sleep a bit to not annoy valve's servers
		time.Sleep(100 * time.Millisecond)
	}
	waitG.Wait()
}

func main() {

	// level := flag.Int("level", 2, "Level of friends you want to crawl. 1 is your friends, 2 is mutual friends etc")
	statMode := flag.Bool("stat", false, "Simple lookup of a target user.")
	testKeys := flag.Bool("testkeys", false, "Test if all keys in APIKEYS.txt are valid")
	flag.Parse()

	apiKeys, err := GetAPIKeys()
	CheckErr(err)

	if *testKeys == true {
		CheckAPIKeys(apiKeys)
		os.Exit(0)
	}

	if len(os.Args) > 1 {
		// Last argument should be the steamID
		controlFunc(apiKeys, os.Args[len(os.Args)-1], *statMode)
	} else {
		fmt.Printf("Incorrect arguments\nUsage: ./main [arguments] steamID\n")
	}
}
