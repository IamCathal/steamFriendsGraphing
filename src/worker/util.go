package worker

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/steamFriendsGraphing/util"
)

// IsEnvVarSet does a simple check to see if an environment
// variable is set
func IsEnvVarSet(envvar string) bool {
	if _, exists := os.LookupEnv(envvar); exists {
		return true
	}
	return false
}

func LoadMappings() map[string]string {
	urlMapLocation := appConfig.UrlMappingsLocation
	if urlMapLocation == "" {
		util.ThrowErr(errors.New("config.UrlMappingsLocation was not initialised before attempting to load url mappings"))
	}
	urlMap := make(map[string]string)
	byteContent, err := ioutil.ReadFile(urlMapLocation)
	if err != nil {
		log.Fatal(err)
	}
	stringContent := string(byteContent)

	if len(stringContent) > 0 {
		lines := strings.Split(stringContent, "\n")

		for _, line := range lines {
			splitArr := strings.Split(line, ":")
			// Last line is just \n
			if len(splitArr) == 2 {
				urlMap[splitArr[0]] = splitArr[1]
			}
		}
		return urlMap

	} else {
		return make(map[string]string)
	}
}

func writeMappings(urlMap map[string]string) {
	urlMapLocation := appConfig.UrlMappingsLocation
	if urlMapLocation == "" {
		util.ThrowErr(errors.New("config.UrlMappingsLocation was not initialised before attempting to write url mappings"))
	}
	file, err := os.OpenFile(urlMapLocation, os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.Seek(0, 0)
	for key, _ := range urlMap {
		_, err = file.WriteString(fmt.Sprintf("%s:%s\n", key, urlMap[key]))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func sortSteamIDs(steamIDs []string) ([]string, error) {
	result := []string{}
	if len(steamIDs) == 1 {
		return append(result, steamIDs[0]), nil
	}
	// Given steamIDs a and b we can create a mapping.
	// Given steamIDs b and a we already have made this graph and
	// don't want to generate a new page
	// Therefore if two steamIDs are given they are first sorted so
	// that given a,b we always create the identifier using a,b and
	// given b,a we always create the identifier using a,b

	steamID1Int64, err := strconv.Atoi(steamIDs[0])
	if err != nil {
		return result, err
	}
	steamID2Int64, err := strconv.Atoi(steamIDs[1])
	if err != nil {
		return result, err
	}

	if steamID1Int64 < steamID2Int64 {
		return []string{strconv.Itoa(steamID1Int64), strconv.Itoa(steamID2Int64)}, nil
	}

	return []string{strconv.Itoa(steamID2Int64), strconv.Itoa(steamID1Int64)}, nil
}

func getSteamIDsIdentifier(steamIDs []string, urlMap map[string]string) (string, error) {
	steamIDs, err := sortSteamIDs(steamIDs)
	return strings.Join(steamIDs, ","), err
}

func GenerateURL(input string, urlMap map[string]string) {
	identifier := ksuid.New()
	urlMap[input] = identifier.String()
	writeMappings(urlMap)
}
