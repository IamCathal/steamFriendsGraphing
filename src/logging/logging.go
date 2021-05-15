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
	appConfig configuration.Info
)

func SetConfig(config configuration.Info) {
	appConfig = config
}

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		path, _ := os.Getwd()

		log.Fatal(fmt.Sprintf(" %s - %s:%d", err, strings.TrimPrefix(file, path), line))
	}
}

func SpecialLog(cntr util.ControllerInterface, msg string) {
	logsFolder := appConfig.LogsFolderLocation
	if logsFolder == "" {
		util.ThrowErr(errors.New("config.LogsFolderLocation was not initialised before attempting to write to file"))
	}

	logg.SetFormatter(&logg.JSONFormatter{})
	file, err := cntr.OpenFile(fmt.Sprintf("%s/%s.txt", logsFolder, os.Getenv("CURRTARGET")), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	CheckErr(err)

	logg.SetOutput(file)
	logg.WithFields(logg.Fields{}).Info(msg)
}
