package upendo

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/solgar/upendo/controller"
	"github.com/solgar/upendo/controller/resources"
	"github.com/solgar/upendo/pages"
	"github.com/solgar/upendo/router"
	"github.com/solgar/upendo/session"
	"github.com/solgar/upendo/settings"
)

// Upendo version variables
var (
	VersionMajor = "0"
	VersionMinor = "1"
)

// Start function starts upendo application with given name
func Start(appName string) {
	settings.Initialize()

	controller.RegisterPreRouteFunctions()
	controller.InstallErrors()
	resources.Install()

	fmt.Printf("upendo ver %s", VersionMajor+"."+VersionMinor)
	fmt.Printf(" running: %s", appName)
	fmt.Println()

	setup()

	fmt.Println("Listening on port", ":"+settings.ServicePort)
	http.HandleFunc("/", router.RouteRequest)
	if settings.CertFile != "" && settings.KeyFile != "" {
		fmt.Println(http.ListenAndServeTLS(":"+settings.ServicePort, settings.CertFile, settings.KeyFile, nil))
	} else {
		fmt.Println(http.ListenAndServe(":"+settings.ServicePort, nil))
	}
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
