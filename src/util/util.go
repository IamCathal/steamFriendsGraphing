package util

import (
	"bufio"
	"compress/gzip"
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

	"github.com/steamFriendsGraphing/configuration"
)

var (
	Green = "\033[32m"
	Red   = "\033[31m"
	White = "\033[0;37m"

	config configuration.Info
)

type Controller struct{}

// ControllerInterface defines all methods that are stubbed for
// service testing due to their dependencies with networks and
// filesystems
type ControllerInterface interface {
	CallPlayerSummaryAPI(steamID, apiKey string) (UserStatsStruct, error)
	CallIsAPIKeyValidAPI(apiKeys string) string
	CallGetFriendsListAPI(steamID, apiKey string) (FriendsStruct, error)

	FileExists(steamID string) bool
	Open(fileName string) (*os.File, error)
	OpenFile(fileName string, flag int, perm os.FileMode) (*os.File, error)
	CreateFile(fileName string) (*os.File, error)
	WriteGzip(file *os.File, content string) error
}

// Open opens a given file and returns a pointer to the file object. Closing
// the file is the responsibility of the function invoking Open
func (controller Controller) Open(fileName string) (*os.File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to open %s", fileName)
		return nil, errors.New(errorMsg)
	}

	return file, nil
}

// Create creates a specified file. Closing the file is the responsibility of
// the function invoking CreateFile
func (controller Controller) CreateFile(fileName string) (*os.File, error) {
	file, err := os.Create(fileName)
	return file, err
}

// WriteGzip gzips a given file
func (controller Controller) WriteGzip(file *os.File, content string) error {
	w := gzip.NewWriter(file)
	_, err := w.Write([]byte(content))
	defer w.Close()

	return err
}

// OpenFile opens a specified with a chosen file mode such as append, write,
// read only etc. Closing the file is the responsibility of the function invoking OpenFile
func (controller Controller) OpenFile(fileName string, flag int, perm os.FileMode) (*os.File, error) {
	file, err := os.OpenFile(fileName, flag, perm)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// CallGetFriendsListAPI calls the Steam GetFriendList API endpoint and returns the response in
// FriendsStruct format
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

// CallPlayerSummaryAPI calls the Steam GetPlayerSummary API endpoint
func (control Controller) CallPlayerSummaryAPI(steamID, apiKey string) (UserStatsStruct, error) {
	var userStatsObj UserStatsStruct
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := GetAndRead(targetURL)
	if err != nil {
		return userStatsObj, err
	}

	json.Unmarshal(res, &userStatsObj)

	return userStatsObj, nil
}

// CallIsAPIKeyValidAPI calls the Steam web API and it's response is used to
// determine if the specified API key is valid
func (control Controller) CallIsAPIKeyValidAPI(apiKey string) string {
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=76561198282036055&relationship=friend", url.QueryEscape(apiKey))
	res, err := http.Get(targetURL)
	CheckErr(err)

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	CheckErr(err)

	return string(body)
}

// FileExists checks is a specified file exists
func (control Controller) FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// SetConfig sets the global config used by various functions
// to manage cache and logging folder file locations
func SetConfig(appConfig configuration.Info) {
	config = appConfig
}

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
	cacheFolder := config.CacheFolderLocation
	if cacheFolder == "" {
		ThrowErr(errors.New("config.CacheFolderLocation was not initialised before attempting to write to file"))
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
	APIKeysLocation := config.ApiKeysFileLocation
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
		return validSteamIDs, fmt.Errorf("No valid steamIDs given")
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
