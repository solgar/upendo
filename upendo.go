package upendo

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"upendo/pages"
	"upendo/router"
	"upendo/session"
	"upendo/settings"

	// blank import to only trigger init
	_ "upendo/controller"
	_ "upendo/controller/resources"
)

// Upendo version variables
var (
	VersionMajor = "0"
	VersionMinor = "1"
)

// Start function starts upendo application with given name
func Start(appName string) {
	fmt.Printf("upendo ver %s", VersionMajor + "." + VersionMinor)
	fmt.Printf(" running: %s", appName)
	fmt.Println()

	setup()

	fmt.Println("Listening on port", ":"+settings.ServicePort)
	http.HandleFunc("/", router.RouteRequest)
	fmt.Println(http.ListenAndServe(":"+settings.ServicePort, nil))
}

// setup is called in main function to start listening for system signals and
// load templates from "templates" folder
func setup() {
	go listenToSignals()
	pages.LoadTemplates(settings.TemplatesDir)
}

// TODO: move it to separate package and parametrize
func listenToSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c

	session.Deinit()

	os.Exit(0)
}
