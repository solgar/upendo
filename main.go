package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"upendo/pages"
	"upendo/router"
	"upendo/session"
	"upendo/settings"

	_ "upendo/controller/loader"
)

// setup is called in main function to start listening for system signals and
// load templates from "templates" folder
func setup() {
	go listenToSignals()
	pages.LoadTemplates("templates")
}

func fullVersion() string {
	return Version +
		" m:" + strconv.Itoa(ModifiedFilesCount) +
		" u:" + strconv.Itoa(UntrackedFilesCount)
}

func main() {
	fmt.Printf("upendo ver %s", fullVersion())
	if Version != BasedOnVersion {
		fmt.Printf(" based on %s", BasedOnVersion)
	}
	fmt.Println()

	setup()

	fmt.Println("Listening on port", ":"+settings.ServicePort)
	http.HandleFunc("/", router.RouteRequest)
	fmt.Println(http.ListenAndServe(":"+settings.ServicePort, nil))
}

// TODO: move it to separate package and parametrize
func listenToSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c

	session.Deinit()

	os.Exit(0)
}
