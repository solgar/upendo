package settings

import (
	"flag"
	"strings"
)

// Customize settings by passing proper parameters. See ./upendo -h for details.
var (
	// can be set to change application starting directory
	StartDir string

	// determines on which port upendo will run
	ServicePort string

	// if set to yes all templates will be reloaded on each request
	ReloadTemplates bool

	// should sessions be stored in file when terminating upendo
	ArchiveSessions bool

	// should sessions be restored from file when launching upendo
	RestoreSessions bool

	// limits chain of routing calls to specified value
	RoutingChainMax int
)

func init() {
	flag.StringVar(&StartDir, "start-dir", "", "app start directory, defaults to \".\"")
	flag.StringVar(&ServicePort, "port", "8080", "port for service to listen on")
	flag.BoolVar(&ReloadTemplates, "reload-templates", false, "if \"true\" then on each request page templates are reloaded")
	flag.BoolVar(&ArchiveSessions, "archive-sessions", true, "if \"true\" upon closing active sessions are archived to file")
	flag.BoolVar(&RestoreSessions, "restore-sessions", true, "if \"true\" restores previously active sessions")
	flag.IntVar(&RoutingChainMax, "routing-chain-max", 4, "limits maximum routing calls to specified value")
	flag.Parse()
	if len(StartDir) > 0 && strings.HasSuffix(StartDir, "/") == false {
		StartDir = StartDir + "/"
	}
}
