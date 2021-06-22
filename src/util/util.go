package util

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/steamFriendsGraphing/configuration"
)

var (
	Green = "\033[32m"
	Red   = "\033[31m"
	White = "\033[0;37m"
)

// GetPlayerSummary gets a player summary through the Steam web API
func GetPlayerSummary(cntr ControllerInterface, steamID, apiKey string) (Player, error) {
	userStatsObj, err := cntr.CallPlayerSummaryAPI(steamID, apiKey)
	if err != nil {
		return Player{}, err
	}

	if len(userStatsObj.Response.Players) == 0 {
		return Player{}, fmt.Errorf("invalid steamID %s given", steamID)
	}

	return userStatsObj.Response.Players[0], nil
}

// GetUsername gets a username from a given steamID by querying the Steam web API
func GetUsername(cntr ControllerInterface, apiKey, steamID string) (string, error) {
	if valid := IsValidFormatSteamID(steamID); !valid {
		return "", fmt.Errorf("invalid steamID format: %s", steamID)
	}
	userStatsObj, err := GetPlayerSummary(cntr, steamID, apiKey)

	return userStatsObj.Personaname, err
}

// GetUserDetails gets profile details such as: steamID, username, time created
// profile URL and avatar URL
func GetUserDetails(cntr ControllerInterface, apiKey, steamID string) (Player, error) {
	userStatsObj, err := GetPlayerSummary(cntr, steamID, apiKey)
	if err != nil {
		return userStatsObj, err
	}

	return userStatsObj, nil
}

// CreateUserDataFolder creates a folder for holding cache.
// Can either be userData for regular use or testData when running under github actions.
func CreateUserDataFolder() error {
	cacheFolder := configuration.AppConfig.CacheFolderLocation
	if cacheFolder == "" {
		return MakeErr(errors.New("configuration.AppConfig.CacheFolderLocation was not initialised before attempting to write to file"))
		// ThrowErr(errors.New("configuration.AppConfig.CacheFolderLocation was not initialised before attempting to write to file"))
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
	_, file, line, _ := runtime.Caller(1)
	path, _ := os.Getwd()
	log.Fatal(fmt.Sprintf("%s:%d ", strings.TrimPrefix(file, path), line), err)
}

func MakeErr(err error, msg ...string) error {
	_, file, line, _ := runtime.Caller(1)
	path, _ := os.Getwd()
	return fmt.Errorf("%s:%d %s %s", strings.TrimPrefix(file, path), line, msg, err)
}

// IsValidFormatSteamID runs a simple regex check to see if the
// steamID is in the valid format before calling the API
func IsValidFormatSteamID(steamID string) bool {
	match, _ := regexp.MatchString("([0-9]){17}", steamID)
	return match
}

// IsValidAPIResponseForSteamId checks if a steamID is valid based
// off of the response from the Steam web API
func IsValidAPIResponseForSteamId(body string) bool {
	match, _ := regexp.MatchString("(Internal Server Error)+", body)
	return !match
}

// IsValidResponseForAPIKey checks if the API key is invalid based
// off of the API response
func IsValidResponseForAPIKey(body string) bool {
	match, _ := regexp.MatchString("(Forbidden)+", body)
	return !match
}

// CheckAPIKeys checks if a given list of API keys is valid by
// calling the Steam web API with each key
func CheckAPIKeys(cntr ControllerInterface, apiKeys []string) error {
	for i, apiKey := range apiKeys {
		response, err := cntr.CallIsAPIKeyValidAPI(apiKey)
		if err != nil {
			return err
		}

		// Wouldn't want to log API keys to console if using
		// the github actions testing environment
		if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
			apiKey = "REDACTED"
		}
		fmt.Printf("[%d] Testing %s ...", i, apiKey)

		if valid := IsValidResponseForAPIKey(response); !valid {
			return MakeErr(fmt.Errorf("invalid api key %s", apiKey))
		}

		fmt.Printf("\r[%d] Testing %s ... %svalid!%s\n", i, apiKey, Green, White)
	}
	fmt.Printf("All API keys are valid!\n")
	return nil
}

// IsEnvVarSet checks if a specified environment variable is set
func IsEnvVarSet(envvar string) bool {
	if _, exists := os.LookupEnv(envvar); exists {
		return true
	}

	return false
}

// GetAPIKeys retrieves the API key(s) to make requests with. API keys must
// be stored in APIKEYS.txt must be saved in the root directory of the projecy
func GetAPIKeys(cntr ControllerInterface) ([]string, error) {
	if exists := IsEnvVarSet("GITHUBACTIONS"); exists {
		return []string{os.Getenv("APIKEY"), os.Getenv("APIKEY1")}, nil
	}

	// APIKEYS.txt MUST be in the root directory of the project
	APIKeysLocation := configuration.AppConfig.ApiKeysFileLocation
	apiKeys := make([]string, 0)

	file, err := cntr.Open(APIKeysLocation)
	if err != nil {
		return apiKeys, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
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

// ExtractSteamIDs returns valid API keys from a specified list and
// returns an error if none are found
func ExtractSteamIDs(args []string) ([]string, error) {
	validSteamIDs := []string{}
	for _, arg := range args {
		if valid := IsValidFormatSteamID(arg); valid {
			validSteamIDs = append(validSteamIDs, arg)
		}
	}
	if len(validSteamIDs) == 0 {
		return validSteamIDs, errors.New("no valid steamIDs given")
	}

	return validSteamIDs, nil
}

// GetAndRead executes a HTTP GET request and returns the body
// of the response in []byte format
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

// AllElementsEmpty determines if all elements in a specified
// list are empty
func AllElementsEmpty(list []string) bool {
	for _, elem := range list {
		if elem != "" {
			return false
		}
	}

	return true
}

// IfKeyNotInMap does what it says on the tin
func IsKeyInUrlMap(key string) bool {
	if _, exists := configuration.AppConfig.UrlMap[key]; exists {
		return true
	}
	return false
}
