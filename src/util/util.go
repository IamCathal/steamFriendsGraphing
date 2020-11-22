package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	Green = "\033[32m"
	// Green color text for terminal printing
	Red = "\033[31m"
	// Red color text for terminal printing
	White = "\033[0;37m"
	// White color text for terminal printing
)

// GetUsername gets a username from a given steamID
func GetUsername(apiKey, steamID string) (string, error) {

	if valid := IsValidFormatSteamID(steamID); !valid {
		return "", fmt.Errorf("invalid steamID format")
	}

	// Get the target username from the ID
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := http.Get(targetURL)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	var userStatsObj UserStatsStruct
	json.Unmarshal(body, &userStatsObj)

	if len(userStatsObj.Response.Players) == 0 {
		return "", fmt.Errorf("invalid steamID %s given", steamID)
	}

	return userStatsObj.Response.Players[0].Personaname, nil
}

func GetUserDetails(apiKey, steamID string) (map[string]string, error) {
	// Get the target username from the ID
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	var userStatsObj UserStatsStruct
	json.Unmarshal(body, &userStatsObj)

	if len(userStatsObj.Response.Players) == 0 {
		return nil, fmt.Errorf("invalid steamID %s given", steamID)
	}

	resMap := make(map[string]string)
	resMap["SteamID"] = userStatsObj.Response.Players[0].Steamid
	resMap["Username"] = userStatsObj.Response.Players[0].Personaname
	resMap["TimeCreated"] = fmt.Sprintf("%s", time.Unix(int64(userStatsObj.Response.Players[0].Timecreated), 0))
	resMap["ProfileURL"] = userStatsObj.Response.Players[0].Profileurl
	resMap["AvatarURL"] = userStatsObj.Response.Players[0].Avatarfull

	return resMap, nil
}

// CreateUserDataFolder creates a folder for holding cache.
// Can either be userData for regular use or testData when running under github actions.
func CreateUserDataFolder() error {
	// Create the cache folder to hold logs if it doesn't exist
	cacheFolder := "userData"
	if exists := IsEnvVarSet("testing"); exists {
		cacheFolder = "testData"
	}

	_, err := os.Stat(fmt.Sprintf("%s", cacheFolder))
	if os.IsNotExist(err) {
		os.Mkdir("userData/", 0755)
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

// IsValidFormatSteamID does a simple regex check to see if the
// steamID is in the valid format before calling the API
func IsValidFormatSteamID(steamID string) bool {
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	if !match {
		return false
	}
	return true
}

// IsValidSteamID checks if a steamID is valid by calling the API
func IsValidSteamID(body string) bool {
	match, _ := regexp.MatchString("(Internal Server Error)+", body)
	if match {
		return false
	}
	return true
}

// IsValidAPIKey checks if the API key is invalid based off of the API
// response
func IsValidAPIKey(body string) bool {
	match, _ := regexp.MatchString("(Forbidden)+", body)
	if match {
		return false
	}
	return true
}

// CheckAPIKeys checks if the API keys in APIKEYS.txt are all valid
func CheckAPIKeys(apiKeys []string) {
	for i, apiKey := range apiKeys {
		targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=76561198282036055&relationship=friend", url.QueryEscape(apiKey))

		// Wouldn't want to log API keys to console if using
		// the github actions testing environment
		if exists := IsEnvVarSet("testing"); exists {
			apiKey = "REDACTED"
		}
		fmt.Printf("[%d] Testing %s ...", i, apiKey)

		res, err := http.Get(targetURL)
		CheckErr(err)
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		CheckErr(err)

		if valid := IsValidAPIKey(string(body)); !valid {
			log.Fatalf("invalid api key %s", apiKey)
		}
		fmt.Printf("\r[%d] Testing %s ... %svalid!%s\n", i, apiKey, Green, White)
		time.Sleep(time.Duration(rand.Intn(1000)+100) * time.Millisecond)
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
func GetAPIKeys() ([]string, error) {
	// APIKEYS.txt MUST be in the root directory of the project
	APIKeysLocation := "../APIKEYS.txt"
	// Dirty fix for now. If testing then go test is invoked in the ./src
	// directory and we should look in the parents parent directory for APIKEYS.txt
	if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
		return []string{os.Getenv("APIKEY"), os.Getenv("APIKEY1")}, nil
	}

	file, err := os.Open(APIKeysLocation)
	if err != nil {
		// Check again in the parent's parent directory if invoked in /src
		file, err = os.Open(fmt.Sprintf("../%s", APIKeysLocation))
		if err != nil {
			CheckErr(errors.New("No APIKEYS.txt file found"))
		}

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

func ExtractSteamIDs(args []string) ([]string, error) {
	validSteamIDs := []string{}
	for _, arg := range args {
		if valid := IsValidFormatSteamID(arg); valid {
			validSteamIDs = append(validSteamIDs, arg)
			fmt.Printf("%s is valid\n", arg)
		}
	}
	if len(validSteamIDs) == 0 {
		return validSteamIDs, fmt.Errorf("No valid steamIDs given")
	}
	return validSteamIDs, nil
}
