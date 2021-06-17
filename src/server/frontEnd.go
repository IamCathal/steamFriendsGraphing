package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func serveGraph(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	finishedGraphLocation := fmt.Sprintf("%s/%s.html", appConfig.FinishedGraphsLocation, vars["id"])
	http.ServeFile(w, req, finishedGraphLocation)
}
