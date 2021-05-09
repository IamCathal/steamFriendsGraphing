package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	logg "github.com/sirupsen/logrus"
	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/util"
)

var (
	config configuration.Info
)

func SetConfig(appConfig configuration.Info) {
	config = appConfig
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		path, _ := os.Getwd()

		log.Fatal(fmt.Sprintf(" %s - %s:%d", err, strings.TrimPrefix(file, path), line))
	}
}

func SpecialLog(msg string) {
	logsFolder := config.LogsFolderLocation
	if logsFolder == "" {
		util.ThrowErr(errors.New("config.LogsFolderLocation was not initialised before attempting to write to file"))
	}

	logg.SetFormatter(&logg.JSONFormatter{})
	f, err := os.OpenFile(fmt.Sprintf("%s/%s.txt", logsFolder, os.Getenv("CURRTARGET")), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	CheckErr(err)

	logg.SetOutput(f)
	logg.WithFields(logg.Fields{}).Info(msg)
}
