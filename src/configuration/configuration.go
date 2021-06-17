package configuration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Info struct {
	CacheFolderLocation     string
	LogsFolderLocation      string
	ApiKeysFileLocation     string
	UrlMappingsLocation     string
	FinishedGraphsLocation  string
	StaticDirectoryLocation string

	IgnoreCache bool
	UrlMap      map[string]string
}

func InitConfig(mode string, dontReadCache bool) Info {
	// baseFolder is the root directory for steamFriendsGraphing
	baseFolder := ""
	cacheFolderLocation := ""
	logsFolderLocation := ""
	apiKeysFileLocation := ""
	urlMappingsLocation := ""
	finishedGraphsLocation := ""
	staticDirectoyLocation := ""

	path, err := os.Getwd()
	CheckErr(err)

	if mode == "testing" {
		baseFolder = fmt.Sprintf("%s/../../", path)
		cacheFolderLocation = filepath.Join(baseFolder, "testData")
		logsFolderLocation = filepath.Join(baseFolder, "testLogs")
		finishedGraphsLocation = filepath.Join(baseFolder, "testFinishedGraphs")
	} else {
		baseFolder = fmt.Sprintf("%s/../", path)
		cacheFolderLocation = filepath.Join(baseFolder, "userData")
		logsFolderLocation = filepath.Join(baseFolder, "logs")
		finishedGraphsLocation = filepath.Join(baseFolder, "finishedGraphs")
	}

	apiKeysFileLocation = filepath.Join(baseFolder, "APIKEYS.txt")
	urlMappingsLocation = filepath.Join(baseFolder, "config/urlMappings.txt")
	staticDirectoyLocation = filepath.Join(baseFolder, "static/")

	return Info{
		CacheFolderLocation:     cacheFolderLocation,
		LogsFolderLocation:      logsFolderLocation,
		ApiKeysFileLocation:     apiKeysFileLocation,
		UrlMappingsLocation:     urlMappingsLocation,
		FinishedGraphsLocation:  finishedGraphsLocation,
		StaticDirectoryLocation: staticDirectoyLocation,
		IgnoreCache:             dontReadCache,
	}
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		path, _ := os.Getwd()
		log.Fatal(fmt.Sprintf("%s:%d ", strings.TrimPrefix(file, path), line), err)
	}
}
