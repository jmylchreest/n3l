package main

import (
	"log"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jmylchreest/n3l/controllers"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/fetch/{host}/{repo}/{group}/{artifact}/{version}/{extension}", controllers.Fetch).Methods("GET")
	router.HandleFunc("/fetch/{host}/{repo}/{group}/{artifact}/{version}/{extension}", controllers.Fetch).Methods("HEAD")
	n := negroni.New(negroni.NewLogger())
	n.UseHandler(router)

	log.Print("N3L, nexus3-latest resolver starting.")
	log.Printf("Version: %s (%s/%s)", GitCommit, GitDescribe, BuildTime)
	n.Run(":3001")
}
