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
	CacheFolderLocation string
	LogsFolderLocation  string
	ApiKeysFileLocation string
}

func InitConfig(mode string) Info {
	// baseFolder is the root directory for steamFriendsGraphing
	baseFolder := ""
	cacheFolderLocation := ""
	logsFolderLocation := ""
	apiKeysFileLocation := ""

	path, err := os.Getwd()
	CheckErr(err)

	if mode == "testing" {
		baseFolder = fmt.Sprintf("%s/../../", path)
		cacheFolderLocation = filepath.Join(baseFolder, "testData")
		logsFolderLocation = filepath.Join(baseFolder, "testLogs")
	} else {
		baseFolder = fmt.Sprintf("%s/../", path)
		cacheFolderLocation = filepath.Join(baseFolder, "userData")
		logsFolderLocation = filepath.Join(baseFolder, "logs")
	}

	apiKeysFileLocation = filepath.Join(baseFolder, "APIKEYS.txt")

	return Info{
		CacheFolderLocation: cacheFolderLocation,
		LogsFolderLocation:  logsFolderLocation,
		ApiKeysFileLocation: apiKeysFileLocation,
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
