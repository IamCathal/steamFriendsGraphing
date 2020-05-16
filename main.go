package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
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

// Divmod divices a friendslist into stacks of 100 and the remainder
func Divmod(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

// LogCall logs a HTTP call to console with HTTP status and round-trip time
func LogCall(method, steamID, username, status, statusColor string, roundTripTime int64) {
	delay := strconv.FormatInt(roundTripTime, 10)
	fmt.Printf("%s [%s] %s %s%s%s %vms\n", method, steamID, username, statusColor, status, "\033[0m", delay)
}

// GetFriends Returns the list of friends of a user in friendsStruct format
func GetFriends(steamID, apiKey string, waitG *sync.WaitGroup) (FriendsStruct, error) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)

	// Check to see if the steamID is in the valid format now to save time
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	if !match {
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		go LogCall("GET", steamID, "\033[31mInvalid SteamID\033[0m", "400", "\033[31m", endTime-startTime)
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

	// If the HTTP response has error messages in it handle them accordingly
	match, _ = regexp.MatchString("(Internal Server Error)+", string(body))
	if match {
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		go LogCall("GET", steamID, friendsObj.Username, "400", "\033[31m", endTime-startTime)
		time.Sleep(1900 * time.Millisecond)
		var temp FriendsStruct
		waitG.Done()
		return temp, errors.New("Invalid steamID given")
	}

	match, _ = regexp.MatchString("(Forbidden)+", string(body))
	if match {
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		go LogCall("GET", steamID, friendsObj.Username, "403", "\033[31m", endTime-startTime)
		time.Sleep(1900 * time.Millisecond)
		var temp FriendsStruct
		waitG.Done()
		return temp, errors.New("Invalid API key -" + apiKey)
	}

	// this part converts the steamIDs we
	// have into usernames that are then added
	// onto the friendsStruct
	steamIDsList := ""
	friendsListLen := len(friendsObj.FriendsList.Friends)

	// Only 100 steamIDs can be given per call, hence we must
	// divide the friends list into lists of 100 or less lists
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
		// More than 100 friends, subsequent calls are needed

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
					// find the entry in the friendsObj struct and set the username field
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
	go LogCall("GET", steamID, friendsObj.Username, "200", "\033[32m", endTime-startTime)
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

// GetAPIKeys Retrieve the API key(s) to make requests with
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
	// APIKeys.txt does exist but it is empty
	return nil, errors.New("No API key(s)")
}

func main() {

	level := flag.Int("level", 2, "Level of friends you want to crawl. 1 is your friends, 2 is mutual friends etc")

	flag.Parse()

	apiKeys, err := GetAPIKeys()
	if err != nil {
		log.Fatal(err)
	}

	numAPIKeys := len(apiKeys)

	// Create the userData folder to hold logs if it doesn't exist
	if _, err := os.Stat("userData/"); os.IsNotExist(err) {
		os.Mkdir("userData/", 0755)
	}

	if len(os.Args) > 1 {
		var waitG sync.WaitGroup
		waitG.Add(1)

		friendsObj, err := GetFriends(os.Args[len(os.Args)-1], apiKeys[0], &waitG)
		if err != nil {
			log.Fatal(err)
		}

		numFriends := len(friendsObj.FriendsList.Friends)

		if *level > 1 {
			fmt.Printf("Friends: %d\nLevels: %d\n", numFriends, *level)
			for i, friend := range friendsObj.FriendsList.Friends {
				waitG.Add(1)
				go GetFriends(friend.Steamid, apiKeys[i%(numAPIKeys)], &waitG)
				// Sleep a bit to not annoy valve's servers
				time.Sleep(100 * time.Millisecond)
			}
			waitG.Wait()
		}

	} else {
		fmt.Println("Incorrect arguments")
		fmt.Println("./main [arguments] steamID")
	}
}
