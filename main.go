package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// IndivFriend holds indivial friend data
type IndivFriend struct {
	Steamid string `json:"steamid"`
	// Steamid is the default steamid64 value for a user
	Relationship string `json:"relationship"`
	// Relationship is always friend in this case
	FriendSince int64 `json:"friend_since,omitempty"`
	// FriendSince, unix timestamp of when the friend request was accepted
	Username string
	// Username is steam username
}

// FriendsStruct messy but it holds the array of all friends
type FriendsStruct struct {
	FriendsList struct {
		Friends []IndivFriend `json:"friends"`
	} `json:"friendslist"`
}

// UserStatsStruct is the JSON response of the
// user stats lookup to get usernames from steamIDs
type UserStatsStruct struct {
	Response struct {
		Players []struct {
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
		} `json:"players"`
	} `json:"response"`
}

// GetFriends Returns the list of friends of a user in friendsStruct format
func GetFriends(steamID, apiKey string) (FriendsStruct, error) {
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	if !match {
		var temp FriendsStruct
		return temp, errors.New("Invalid steamID")
	}

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s&relationship=friend", apiKey, steamID)
	res, err := http.Get(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var friendsObj FriendsStruct
	json.Unmarshal(body, &friendsObj)

	match, _ = regexp.MatchString("(Internal Server Error)+", string(body))
	if match {
		var temp FriendsStruct
		return temp, errors.New("Invalid steamID given")
	}

	match, _ = regexp.MatchString("(Forbidden)+", string(body))
	if match {
		var temp FriendsStruct
		return temp, errors.New("Invalid API key")
	}

	// this part converts the steamIDs we
	// have into usernames that are readable

	steamIDsList := ""
	friendsListLen := len(friendsObj.FriendsList.Friends)
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
	if err != nil {
		log.Fatal(err)
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var userStatsObj UserStatsStruct
	json.Unmarshal(body, &userStatsObj)

	for i, user := range userStatsObj.Response.Players {
		friendsObj.FriendsList.Friends[i].Username = user.Personaname
	}

	WriteToFile(apiKey, steamID, friendsObj)

	return friendsObj, nil
}

// WriteToFile writes the friends to a file for later processing
func WriteToFile(apiKey, steamID string, friends FriendsStruct) {

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamID)
	res, err := http.Get(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var userStatsObj UserStatsStruct
	json.Unmarshal(body, &userStatsObj)

	fileLoc := fmt.Sprintf("userData/%s.txt", userStatsObj.Response.Players[0].Personaname)
	file, err := os.Create(fileLoc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, val := range friends.FriendsList.Friends {
		file.WriteString(val.Steamid + "\t" + val.Username + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

// PrintDetails Returns a nicely formatted array of each friend's name
// and the time since the use first friended them
func PrintDetails(inputObj FriendsStruct) {
	for ind, val := range inputObj.FriendsList.Friends {
		// fmt.Printf("[%d] %v\n", ind+1, time.Unix(val.FriendSince, 0))
		fmt.Printf("[%d] %v\n", ind+1, val.Steamid)
	}
}

// GetAPIKeys Retrieve the API keys to make requests with
func GetAPIKeys() ([]string, error) {
	file, err := os.Open("APIKEYS.txt")
	if err != nil {
		return nil, errors.New("No APIKEYS.txt file found")
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

	return nil, errors.New("No API key(s)")
}

func main() {
	apiKeys, err := GetAPIKeys()
	if err != nil {
		log.Fatal(err)
	}
	level, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("Invalid level entered")
	}

	if _, err := os.Stat("userData/"); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.Mkdir("userData/", 0755)
	}

	if len(os.Args) > 0 {
		friendsObj, err := GetFriends(os.Args[1], apiKeys[0])
		if err != nil {
			log.Fatal(err)
		}
		numFriends := len(friendsObj.FriendsList.Friends)
		if level > 1 {
			fmt.Printf("You have %d friends, this should take %d seconds to scrape at %d level deep (friends of friends)\n", numFriends, numFriends*2, level)
			for i, friend := range friendsObj.FriendsList.Friends {
				_, err := GetFriends(friend.Steamid, apiKeys[i])
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Using key " + apiKeys[i])
			}
		}

	} else {
		fmt.Println("Incorrect arguments")
		fmt.Println("./main steamID")
	}
}
