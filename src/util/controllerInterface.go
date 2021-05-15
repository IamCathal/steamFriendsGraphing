package util

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	os "os"
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

// CallPlayerSummaryAPI calls the Steam GetPlayerSummary API endpoint
func (control Controller) CallPlayerSummaryAPI(steamID, apiKey string) (UserStatsStruct, error) {
	var userStatsObj UserStatsStruct
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamID)
	res, err := GetAndRead(targetURL)
	if err != nil {
		return userStatsObj, MakeErr(err)
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

// CallGetFriendsListAPI calls the Steam GetFriendList API endpoint and returns the response in
// FriendsStruct format
func (controller Controller) CallGetFriendsListAPI(steamID, apiKey string) (FriendsStruct, error) {
	var friendsObj FriendsStruct
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s&relationship=friend", url.QueryEscape(apiKey), url.QueryEscape(steamID))
	res, err := http.Get(targetURL)
	if err != nil {
		return friendsObj, MakeErr(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return friendsObj, MakeErr(err)
	}

	if valid := IsValidAPIResponseForSteamId(string(body)); !valid {
		return friendsObj, MakeErr(fmt.Errorf("invalid steamID %s given", steamID))
	}

	if valid := IsValidResponseForAPIKey(string(body)); !valid {
		return friendsObj, MakeErr(fmt.Errorf("invalid api key: %s", apiKey))
	}

	json.Unmarshal(body, &friendsObj)

	return friendsObj, nil
}

// FileExists checks is a specified file exists
func (control Controller) FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// Open opens a given file and returns a pointer to the file object. Closing
// the file is the responsibility of the function invoking Open
func (controller Controller) Open(fileName string) (*os.File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, MakeErr(fmt.Errorf("failed to open %s", fileName))
	}

	return file, nil
}

// OpenFile opens a specified with a chosen file mode such as append, write,
// read only etc. Closing the file is the responsibility of the function invoking OpenFile
func (controller Controller) OpenFile(fileName string, flag int, perm os.FileMode) (*os.File, error) {
	file, err := os.OpenFile(fileName, flag, perm)
	if err != nil {
		return nil, MakeErr(err)
	}

	return file, nil
}

// Create creates a specified file. Closing the file is the responsibility of
// the function invoking CreateFile
func (controller Controller) CreateFile(fileName string) (*os.File, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return file, MakeErr(err)
	}
	return file, nil
}

// WriteGzip gzips a given file
func (controller Controller) WriteGzip(file *os.File, content string) error {
	w := gzip.NewWriter(file)
	_, err := w.Write([]byte(content))
	defer w.Close()

	fmt.Println("controllerIntercace.WriteGzip: Error writing gzip")
	return err
}
