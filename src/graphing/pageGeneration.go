package graphing

import (
	"fmt"
	"html/template"

	"github.com/steamFriendsGraphing/configuration"
	"github.com/steamFriendsGraphing/util"
)

type graphPageInfo struct {
	GraphCode string
	ID        string
}

func GenerateGraphPage(cntr util.ControllerInterface, ID string) error {
	graphData := graphPageInfo{
		GraphCode: "",
		ID:        ID,
	}
	fullFilename := fmt.Sprintf("%s/%s.html", configuration.AppConfig.FinishedGraphsLocation, ID)
	templateLocation := fmt.Sprintf("%s/graphPage.html", configuration.AppConfig.TemplateDirectory)

	tmpl, err := template.ParseFiles(templateLocation)

	file, err := cntr.CreateFile(fullFilename)
	if err != nil {
		return util.MakeErr(err)
	}

	err = tmpl.Execute(file, graphData)
	if err != nil {
		return util.MakeErr(err)
	}

	file.Close()
	return nil
}
