package logging

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	logg "github.com/sirupsen/logrus"
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

func SpecialLog(msg string) {
	logsFolder := "../logs"
	if exists := util.IsEnvVarSet("testing"); exists {
		logsFolder = "../testLogs"
	}

	logg.SetFormatter(&logg.JSONFormatter{})
	f, err := os.OpenFile(fmt.Sprintf("%s/%s.txt", logsFolder, os.Getenv("CURRTARGET")), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	CheckErr(err)
	// endTime := time.Now().UnixNano() / int64(time.Millisecond)
	// delay := strconv.FormatInt((endTime - startTime), 10)

	logg.SetOutput(f)
	logg.WithFields(logg.Fields{}).Info(msg)
}
