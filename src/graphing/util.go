package graphing

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/steamFriendsGraphing/util"
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
	cacheFolderLocation := appConfig.CacheFolderLocation
	if cacheFolderLocation == "" {
		return temp, util.MakeErr(errors.New("appConfig.CacheFolderLocation was not initialised before attempting to write graph to file"))
	}
	file, err := os.Open(fmt.Sprintf("%s/%s.gz", cacheFolderLocation, steamID))
	if err != nil {
		return FriendsStruct{}, util.MakeErr(err)
	}
	defer file.Close()

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
	cacheFolderLocation := appConfig.CacheFolderLocation
	if cacheFolderLocation == "" {
		return "", util.MakeErr(errors.New("appConfig.CacheFolderLocation was not initialised before attempting to get username from cache file"))
	}
	file, err := os.Open(fmt.Sprintf("%s/%s.gz", cacheFolderLocation, steamID))
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

func NodeExistsInt(ID int, nodeMap map[int]bool) bool {
	_, ok := nodeMap[ID]
	if ok {
		return true
	}
	return false
}

// CreateUserDataFolder creates a folder for holding cache.
// Can either be userData for regular use or testData when running under github actions.
func CreateFinishedGraphFolder() error {
	finishedGraphsLocation := appConfig.FinishedGraphsLocation
	if finishedGraphsLocation == "" {
		return util.MakeErr(errors.New("appConfig.finishedGraphsLocation was not initialised before attempting to create finished graphs folder"))
	}
	_, err := os.Stat(finishedGraphsLocation)
	if os.IsNotExist(err) {
		os.Mkdir(finishedGraphsLocation, 0755)
		return nil
	}
	if err != nil {
		return util.MakeErr(err)
	}
	return nil
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		path, _ := os.Getwd()

		log.Fatal(fmt.Sprintf(" %s - %s:%d", err, strings.TrimPrefix(file, path), line))
	}
}

// GetKeyFromValue gets a key for a given map value
func GetKeyFromValue(userMap map[int]string, value string) (int, bool) {
	for key, val := range userMap {
		if val == value {
			return key, true
		}
	}
	return -1, false
}

// MergeNodes merges the node lists of the starting and target users. If there are
// duplicate nodes in the graphing stage then the graphing framework will fail.
func MergeNodes(firstNodes, secondNodes []charts.GraphNode) []charts.GraphNode {
	foundNodes := make(map[string]bool)
	allNodes := make([]charts.GraphNode, 0)

	secondUsername := secondNodes[0].Name

	for _, node := range firstNodes {
		if node.Name != secondUsername {
			allNodes = append(allNodes, node)
			foundNodes[node.Name] = true
			// fmt.Printf("[%d : %s]\n", len(allNodes), node.Name)
		}
	}

	for _, node := range secondNodes {
		if _, existing := foundNodes[node.Name]; !existing {
			allNodes = append(allNodes, node)
			// fmt.Printf("[%d : %s]\n", len(allNodes), node.Name)
		}
	}
	return allNodes
}
