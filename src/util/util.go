package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	Green = "\033[32m"
	Red   = "\033[31m"
	White = "\033[0;37m"
)

type Controller struct{}

type ControllerInterface interface {
	CallPlayerSummaryAPI(steamID, apiKey string) (UserStatsStruct, error)
	CallIsAPIKeyValidAPI(apiKeys string) string
	CallGetFriendsListAPI(steamID, apiKey string) (FriendsStruct, error)

	FileExists(steamID string) bool
	OpenFile(fileName string) (*bufio.Scanner, error)
}

func (controller Controller) OpenFile(fileName string) (*bufio.Scanner, error) {
	file, err := os.Open(fileName)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to open %s", fileName)
		return nil, errors.New(errorMsg)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	return scanner, nil
}

func (controller Controller) CallGetFriendsListAPI(steamID, apiKey string) (FriendsStruct, error) {
	var friendsObj FriendsStruct
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s&relationship=friend", url.QueryEscape(apiKey), url.QueryEscape(steamID))
	res, err := http.Get(targetURL)
	if err != nil {
		return friendsObj, err
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return friendsObj, err
	}

	if valid := IsValidAPIResponseForSteamId(string(body)); !valid {
		return friendsObj, fmt.Errorf("invalid steamID %s given", steamID)
	}

	if valid := IsValidResponseForAPIKey(string(body)); !valid {
		return friendsObj, fmt.Errorf("invalid api key: %s", apiKey)
	}

	json.Unmarshal(body, &friendsObj)
	return friendsObj, nil
}

func (control Controller) CallPlayerSummaryAPI(steamID, apiKey string) (UserStatsStruct, error) {
	var userStatsObj UserStatsStruct
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := GetAndRead(targetURL)
	if err != nil {
		return userStatsObj, err
	}

	fmt.Println("REMOTE REMOTE REMOTE")
	json.Unmarshal(res, &userStatsObj)
	return userStatsObj, nil
}

func (control Controller) FileExists(steamID string) bool {
	cacheFolder := ""
	if exists := IsEnvVarSet("testing"); exists {
		cacheFolder = fmt.Sprintf("%stestData", os.Getenv("BWD"))
	} else {
		cacheFolder = fmt.Sprintf("%suserData", os.Getenv("BWD"))
	}

	_, err := os.Stat(fmt.Sprintf("%s/%s.gz", cacheFolder, steamID))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// GetPlayerSummary gets a player summary through the steam web API
func GetPlayerSummary(cntr ControllerInterface, steamID, apiKey string) (UserStatsStruct, error) {
	userStatsObj, err := cntr.CallPlayerSummaryAPI(steamID, apiKey)
	if err != nil {
		return userStatsObj, err
	}

	if len(userStatsObj.Response.Players) == 0 {
		return userStatsObj, fmt.Errorf("invalid steamID %s given", steamID)
	}

	return userStatsObj, nil
}

// GetUsername gets a username from a given steamID by querying the
// steam web API
func GetUsername(cntr ControllerInterface, apiKey, steamID string) (string, error) {
	if valid := IsValidFormatSteamID(steamID); !valid {
		return "", fmt.Errorf("invalid steamID format: %s", steamID)
	}
	userStatsObj, err := GetPlayerSummary(cntr, steamID, apiKey)
	return userStatsObj.Response.Players[0].Personaname, err
}

// GetUserDetails gets profile details such as: steamID, username, time created
// profile URL and avatar URL
func GetUserDetails(cntr ControllerInterface, apiKey, steamID string) (map[string]string, error) {
	userStatsObj, err := GetPlayerSummary(cntr, steamID, apiKey)
	if err != nil {
		return nil, err
	}

	resMap := make(map[string]string)
	resMap["SteamID"] = userStatsObj.Response.Players[0].Steamid
	resMap["Username"] = userStatsObj.Response.Players[0].Personaname
	resMap["TimeCreated"] = fmt.Sprintf("%s", time.Unix(int64(userStatsObj.Response.Players[0].Timecreated), 0))
	resMap["ProfileURL"] = userStatsObj.Response.Players[0].Profileurl
	resMap["AvatarURL"] = userStatsObj.Response.Players[0].Avatarfull
	return resMap, err
}

// CreateUserDataFolder creates a folder for holding cache.
// Can either be userData for regular use or testData when running under github actions.
func CreateUserDataFolder() error {
	cacheFolder := ""
	if exists := IsEnvVarSet("testing"); exists {
		cacheFolder = fmt.Sprintf("%s/testData", os.Getenv("BWD"))
	} else {
		cacheFolder = fmt.Sprintf("%s/userData", os.Getenv("BWD"))
	}

	if _, err := os.Stat(cacheFolder); os.IsNotExist(err) {
		err = os.Mkdir(cacheFolder, 0755)
		CheckErr(err)
	}

	return nil
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		path, _ := os.Getwd()
		log.Fatal(fmt.Sprintf("%s:%d ", strings.TrimPrefix(file, path), line), err)
	}
}

// ThrowErr throws an error explicitly
func ThrowErr(err error) {
	CheckErr(err)
}

// IsValidFormatSteamID runs a simple regex check to see if the
// steamID is in the valid format before calling the API
func IsValidFormatSteamID(steamID string) bool {
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	return match
}

// IsValidAPIResponseForSteamId checks if a steamID is valid by calling the API
func IsValidAPIResponseForSteamId(body string) bool {
	match, _ := regexp.MatchString("(Internal Server Error)+", body)
	return !match
}

// IsValidResponseForAPIKey checks if the API key is invalid based off of the API
// response
func IsValidResponseForAPIKey(body string) bool {
	match, _ := regexp.MatchString("(Forbidden)+", body)
	return !match
}

func (control Controller) CallIsAPIKeyValidAPI(apiKey string) string {
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=76561198282036055&relationship=friend", url.QueryEscape(apiKey))
	res, err := http.Get(targetURL)
	CheckErr(err)

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	CheckErr(err)
	return string(body)
}

// CheckAPIKeys checks if a given list of API keys is valid
func CheckAPIKeys(cntr ControllerInterface, apiKeys []string) {
	for i, apiKey := range apiKeys {
		response := cntr.CallIsAPIKeyValidAPI(apiKey)

		// Wouldn't want to log API keys to console if using
		// the github actions testing environment
		if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
			apiKey = "REDACTED"
		}
		fmt.Printf("[%d] Testing %s ...", i, apiKey)

		if valid := IsValidResponseForAPIKey(response); !valid {
			ThrowErr(fmt.Errorf("invalid api key %s", apiKey))
		}

		fmt.Printf("\r[%d] Testing %s ... %svalid!%s\n", i, apiKey, Green, White)
	}
	fmt.Printf("All API keys are valid!\n")
}

func IsEnvVarSet(envvar string) bool {
	if _, exists := os.LookupEnv(envvar); exists {
		return true
	}
	return false
}

// GetAPIKeys retrieves the API key(s) to make requests with
// API keys must be stored in APIKEY(s).txt
func GetAPIKeys(cntr ControllerInterface) ([]string, error) {
	// Dirty fix for now. If testing then go test is invoked in the ./src
	// directory and we should look in the parents parent directory for APIKEYS.txt
	if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
		return []string{os.Getenv("APIKEY"), os.Getenv("APIKEY1")}, nil
	}

	// APIKEYS.txt MUST be in the root directory of the project
	APIKeysLocation := fmt.Sprintf("%s/../APIKEYS.txt", os.Getenv("BWD"))
	apiKeys := make([]string, 0)

	scanner, err := cntr.OpenFile(APIKeysLocation)
	if err != nil {
		return apiKeys, err
	}
	for scanner.Scan() {
		apiKeys = append(apiKeys, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading APIKEYS.txt: %s", err.Error())
	}
	if empty := AllElementsEmpty(apiKeys); empty {
		return nil, errors.New("APIKEYS.txt exists but has no API key(s)")
	}
	return apiKeys, nil

}

func ExtractSteamIDs(args []string) ([]string, error) {
	validSteamIDs := []string{}
	for _, arg := range args {
		if valid := IsValidFormatSteamID(arg); valid {
			validSteamIDs = append(validSteamIDs, arg)
		}
	}
	if len(validSteamIDs) == 0 {
		return validSteamIDs, fmt.Errorf("No valid steamIDs given")
	}

	return validSteamIDs, nil
}

func SetBaseWorkingDirectory() {
	path, err := os.Getwd()
	CheckErr(err)
	os.Setenv("BWD", path)
}

func GetAndRead(URL string) ([]byte, error) {
	res, err := http.Get(URL)
	if err != nil {
		return []byte{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func AllElementsEmpty(list []string) bool {
	for _, elem := range list {
		if elem != "" {
			return false
		}
	}
	return true
}
