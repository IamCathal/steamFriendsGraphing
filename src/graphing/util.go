package graphing

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

type infoStruct struct {
	level    int
	steamID  string
	username string
	from     string
}

// GetCache gets a user's cached records if it exists
func GetCache(steamID string) (FriendsStruct, error) {
	var temp FriendsStruct
	cacheFolder := "userData"
	if exists := IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}
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

// IsEnvVarSet does a simple check to see if an environment
// variable is set
func IsEnvVarSet(envvar string) bool {
	if _, exists := os.LookupEnv(envvar); exists {
		return true
	}
	return false
}

// GetUsernameFromCacheFile gets the username for a given cache file
// e.g 76561198063271448 -> moose
func GetUsernameFromCacheFile(steamID string) (string, error) {
	var temp FriendsStruct
	cacheFolder := "userData"
	if exists := IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}
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

// NodeExists checks if a given node has been added to the graph yet
func NodeExists(username string, nodeMap map[string]bool) bool {
	_, ok := nodeMap[username]
	if ok {
		return true
	}
	return false
}

// CreateUserDataFolder creates a folder for holding cache.
// Can either be userData for regular use or testData when running under github actions.
func CreateFinishedGraphFolder() error {

	_, err := os.Stat("../finishedGraphs")
	if os.IsNotExist(err) {
		os.Mkdir("../finishedGraphs", 0755)
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
