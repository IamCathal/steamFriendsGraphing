package worker

import (
	"os"
	"strconv"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/steamFriendsGraphing/configuration"
)

// IsEnvVarSet does a simple check to see if an environment
// variable is set
func IsEnvVarSet(envvar string) bool {
	if _, exists := os.LookupEnv(envvar); exists {
		return true
	}
	return false
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
	configuration.WriteMappings(appConfig, urlMap)
}
