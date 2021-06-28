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
	TemplateDirectory       string

	// Configuration flags
	IgnoreCache bool
	AlwaysCrawl bool

	UrlMap map[string]string
}

var (
	AppConfig Info
)

func SetConfig(config Info) {
	AppConfig = config
}

func InitAndSetConfig(mode string, dontReadCache, alwaysCrawl bool) {
	// baseFolder is the root directory for steamFriendsGraphing
	baseFolder := ""
	cacheFolderLocation := ""
	logsFolderLocation := ""
	apiKeysFileLocation := ""
	urlMappingsLocation := ""
	finishedGraphsLocation := ""
	staticDirectoyLocation := ""
	templateDirectory := ""

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
		finishedGraphsLocation = filepath.Join(baseFolder, "static/graph")
	}

	apiKeysFileLocation = filepath.Join(baseFolder, "APIKEYS.txt")
	urlMappingsLocation = filepath.Join(baseFolder, "config/urlMappings.txt")
	staticDirectoyLocation = filepath.Join(baseFolder, "static/")
	templateDirectory = filepath.Join(baseFolder, "/templates")

	initialisedAppConfig := Info{
		CacheFolderLocation:     cacheFolderLocation,
		LogsFolderLocation:      logsFolderLocation,
		ApiKeysFileLocation:     apiKeysFileLocation,
		UrlMappingsLocation:     urlMappingsLocation,
		FinishedGraphsLocation:  finishedGraphsLocation,
		StaticDirectoryLocation: staticDirectoyLocation,
		TemplateDirectory:       templateDirectory,
		IgnoreCache:             dontReadCache,
		AlwaysCrawl:             alwaysCrawl,
	}

	urlMap := make(map[string]string)

	if mode != "testing" {
		urlMap, err = loadMappings(initialisedAppConfig)
		CheckErr(err)
	}
	initialisedAppConfig.UrlMap = urlMap
	SetConfig(initialisedAppConfig)
}

func loadMappings(underConstructionConfig Info) (map[string]string, error) {
	urlMapLocation := underConstructionConfig.UrlMappingsLocation
	if urlMapLocation == "" {
		return nil, MakeErr(errors.New("underConstructionConfig.UrlMappingsLocation was not initialised before attempting to load url mappings"))
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

func WriteMappings() error {
	urlMapLocation := AppConfig.UrlMappingsLocation
	if urlMapLocation == "" {
		return MakeErr(errors.New("appConfig.UrlMappingsLocation was not initialised before attempting to write url mappings"))
	}
	file, err := os.OpenFile(urlMapLocation, os.O_RDWR, 0755)
	if err != nil {
		return MakeErr(err)
	}
	defer file.Close()
	file.Seek(0, 0)
	for key, _ := range AppConfig.UrlMap {
		_, err = file.WriteString(fmt.Sprintf("%s:%s\n", key, AppConfig.UrlMap[key]))
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
