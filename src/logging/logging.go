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

// CheckErr is a simple function to replace dozen or so if err != nil statements
func CheckErr(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		path, _ := os.Getwd()

		log.Fatal(fmt.Sprintf(" %s - %s:%d", err, strings.TrimPrefix(file, path), line))
	}
}

// SpecialLog logs to file using logrus
func SpecialLog(cntr util.ControllerInterface, logFileName, msg string) error {
	logsFolder := configuration.AppConfig.LogsFolderLocation
	if logsFolder == "" {
		return util.MakeErr(errors.New("configuration.AppConfig.LogsFolderLocation was not initialised before attempting to write to file"))
	}

	logg.SetFormatter(&logg.JSONFormatter{})
	file, err := cntr.OpenFile(fmt.Sprintf("%s/%s.txt", logsFolder, logFileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return util.MakeErr(err)
	}

	logg.SetOutput(file)
	logg.WithFields(logg.Fields{}).Info(msg)
	return nil
}
