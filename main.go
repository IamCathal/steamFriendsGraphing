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
	"sync"
	"time"
)

// IndivFriend holds indivial friend data
type IndivFriend struct {
	Steamid string `json:"steamid"`
	// Steamid is the default steamid64 value for a user
	Relationship string `json:"relationship"`
	// Relationship is always friend in this case
	FriendSince int64 `json:"friend_since,omitempty"`
	// FriendSince, unix timestamp of when the friend request was accepted
	Username string `json:"username"`
	// Username is steam username
}

// FriendsStruct messy but it holds the array of all friends
type FriendsStruct struct {
	Username    string `json:"username"`
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

func divmod(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

func logCall(method, steamID, username, status, statusColor string, roundTripTime int64) {
	delay := strconv.FormatInt(roundTripTime, 10)
	fmt.Printf("%s [%s] %s %s%s%s %vms\n", method, steamID, username, statusColor, status, "\033[0m", delay)
}

// GetFriends Returns the list of friends of a user in friendsStruct format
func GetFriends(steamID, apiKey string, waitG *sync.WaitGroup) (FriendsStruct, error) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	if !match {
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		go logCall("GET", steamID, "\033[31mInvalid SteamID\033[0m", "400", "\033[31m", endTime-startTime)
		var temp FriendsStruct
		waitG.Done()
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
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		go logCall("GET", steamID, friendsObj.Username, "400", "\033[31m", endTime-startTime)
		time.Sleep(1900 * time.Millisecond)
		var temp FriendsStruct
		waitG.Done()
		return temp, errors.New("Invalid steamID given")
	}

	match, _ = regexp.MatchString("(Forbidden)+", string(body))
	if match {
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		go logCall("GET", steamID, friendsObj.Username, "403", "\033[31m", endTime-startTime)
		time.Sleep(1900 * time.Millisecond)
		var temp FriendsStruct
		waitG.Done()
		return temp, errors.New("Invalid API key -" + apiKey)
	}

	// this part converts the steamIDs we
	// have into usernames that are readable

	steamIDsList := ""
	friendsListLen := len(friendsObj.FriendsList.Friends)

	callCount, remainder := divmod(friendsListLen, 100)

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

	} else {
		// divide into calls of 100 friends and usernames to the friends struct

		for i := 0; i <= callCount; i++ {
			//each batch of 100

			steamIDsList = ""

			if i < callCount {
				// a full call of 100 friends
				for k := 0; k < 100; k++ {
					// fmt.Println(k + (i * 100))
					steamIDsList += friendsObj.FriendsList.Friends[k+(i*100)].Steamid + ","
				}
			} else {
				// a batch of the remainder (less than 100)
				for k := 0; k < remainder; k++ {
					steamIDsList += friendsObj.FriendsList.Friends[k+(i*100)].Steamid + ","
					// fmt.Println(k + (i * 100))
				}
			}

			targetURL = fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamIDsList)
			// fmt.Println(targetURL)
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

			if i < callCount {
				for k := 0; k < 100; k++ {
					friendsObj.FriendsList.Friends[k+(i*100)].Username = userStatsObj.Response.Players[k].Personaname
				}
			} else {
				for k := 0; k < remainder; k++ {
					friendsObj.FriendsList.Friends[k+(i*100)].Username = userStatsObj.Response.Players[k].Personaname
				}
			}

		}

	}

	// get the target person's username
	targetURL = fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", apiKey, steamID)
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

	friendsObj.Username = userStatsObj.Response.Players[0].Personaname

	// if testing env is set, don't bother writing to file
	if os.Getenv("testing") == "" {
		WriteToFile(apiKey, steamID, friendsObj)
	}

	// log the request along the round trip delay
	endTime := time.Now().UnixNano() / int64(time.Millisecond)
	go logCall("GET", steamID, friendsObj.Username, "200", "\033[32m", endTime-startTime)
	waitG.Done()
	return friendsObj, nil
}

// WriteToFile writes the friends to a file for later processing
func WriteToFile(apiKey, steamID string, friends FriendsStruct) {

	fileLoc := fmt.Sprintf("userData/%s.json", steamID)
	file, err := os.Create(fileLoc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	jsonObj, err := json.Marshal(friends)
	if err != nil {
		log.Fatal(err)
	}

	_ = ioutil.WriteFile(fileLoc, jsonObj, 0644)
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
	if len(os.Args) < 3 {
		log.Fatal("Invalid arguments entered")
	}

	apiKeys, err := GetAPIKeys()
	if err != nil {
		log.Fatal(err)
	}

	numAPIKeys := len(apiKeys)

	level, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("Invalid level entered")
	}

	if _, err := os.Stat("userData/"); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.Mkdir("userData/", 0755)
	}

	if len(os.Args) > 0 {

		var waitG sync.WaitGroup
		waitG.Add(1)
		friendsObj, err := GetFriends(os.Args[1], apiKeys[0], &waitG)
		if err != nil {
			log.Fatal(err)
		}
		numFriends := len(friendsObj.FriendsList.Friends)
		if level > 1 {
			fmt.Printf("You have %d friends, this should take %d seconds to scrape at %d level deep (friends of friends)\n", numFriends, numFriends*2, level)
			for i, friend := range friendsObj.FriendsList.Friends {
				// _, err := GetFriends(friend.Steamid, apiKeys[i%(numAPIKeys)])
				waitG.Add(1)
				go GetFriends(friend.Steamid, apiKeys[i%(numAPIKeys)], &waitG)
				time.Sleep(100 * time.Millisecond)

				// if err != nil {
				// 	log.Fatal(err)
				// }
				// fmt.Println("Using key " + apiKeys[i])
			}

			waitG.Wait()
		}

	} else {
		fmt.Println("Incorrect arguments")
		fmt.Println("./main steamID")
	}
}
