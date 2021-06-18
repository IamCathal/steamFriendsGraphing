package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
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

	initialisedAppConfig := Info{
		CacheFolderLocation:     cacheFolderLocation,
		LogsFolderLocation:      logsFolderLocation,
		ApiKeysFileLocation:     apiKeysFileLocation,
		UrlMappingsLocation:     urlMappingsLocation,
		FinishedGraphsLocation:  finishedGraphsLocation,
		StaticDirectoryLocation: staticDirectoyLocation,
		IgnoreCache:             dontReadCache,
	}

	urlMap, err := loadMappings(initialisedAppConfig)
	CheckErr(err)

	initialisedAppConfig.UrlMap = urlMap

	return initialisedAppConfig
}

func loadMappings(appConfig Info) (map[string]string, error) {
	urlMapLocation := appConfig.UrlMappingsLocation
	if urlMapLocation == "" {
		return nil, MakeErr(errors.New("appConfig.UrlMappingsLocation was not initialised before attempting to load url mappings"))
	}
	urlMap := make(map[string]string)
	byteContent, err := ioutil.ReadFile(urlMapLocation)
	if err != nil {
		return nil, MakeErr(err)
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
		return urlMap, nil

	} else {
		return make(map[string]string), nil
	}
}

func WriteMappings(appConfig Info, urlMap map[string]string) error {
	urlMapLocation := appConfig.UrlMappingsLocation
	if urlMapLocation == "" {
		return MakeErr(errors.New("appConfig.UrlMappingsLocation was not initialised before attempting to write url mappings"))
	}
	file, err := os.OpenFile(urlMapLocation, os.O_RDWR, 0755)
	if err != nil {
		return MakeErr(err)
	}
	defer file.Close()
	file.Seek(0, 0)
	for key, _ := range urlMap {
		_, err = file.WriteString(fmt.Sprintf("%s:%s\n", key, urlMap[key]))
		if err != nil {
			return MakeErr(err)
		}
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

func MakeErr(err error, msg ...string) error {
	_, file, line, _ := runtime.Caller(1)
	path, _ := os.Getwd()
	return fmt.Errorf("%s:%d %s %s", strings.TrimPrefix(file, path), line, msg, err)
}
