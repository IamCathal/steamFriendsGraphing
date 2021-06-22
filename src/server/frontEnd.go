package server

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/steamFriendsGraphing/configuration"
)

func serveGraph(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	finishedGraphLocation := fmt.Sprintf("%s/%s.html", configuration.AppConfig.FinishedGraphsLocation, path.Clean(vars["id"]))
	http.ServeFile(w, req, finishedGraphLocation)
}
