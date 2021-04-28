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

	// panic if templayes cannot be found
	RequireTemplates bool

	// default relative location where to look for templates
	TemplatesDir string

	// should sessions be stored in file when terminating upendo
	ArchiveSessions bool

	// should sessions be restored from file when launching upendo
	RestoreSessions bool

	// limits chain of routing calls to specified value
	RoutingChainMax int

	// if map files are ignored (js.map, css.map)
	IgnoreMapFiles bool

	// cert file
	CertFile string

	// key file
	KeyFile string

	// not implemented yet
	LoadSettingsFromFile bool
)

func init() {
	flag.StringVar(&StartDir, "start-dir", "", "app start directory, defaults to \".\"")
	flag.StringVar(&ServicePort, "port", "8080", "port for service to listen on")
	flag.BoolVar(&ReloadTemplates, "reload-templates", false, "if \"true\" then on each request page templates are reloaded")
	flag.BoolVar(&RequireTemplates, "require-templates", false, "if \"true\" then panic if no templates can be found, ignore otherwise")
	flag.StringVar(&TemplatesDir, "templates-dir", "templates", "default relative location to look for templates")
	flag.BoolVar(&ArchiveSessions, "archive-sessions", true, "if \"true\" upon closing active sessions are archived to file")
	flag.BoolVar(&RestoreSessions, "restore-sessions", true, "if \"true\" restores previously active sessions")
	flag.BoolVar(&IgnoreMapFiles, "ignore-map-files", true, "if \"true\" \"file not found\" errors for .map files will be ignored")
	flag.IntVar(&RoutingChainMax, "routing-chain-max", 4, "limits maximum routing calls to specified value")
	flag.BoolVar(&LoadSettingsFromFile, "settings-from-file", false, "if \"true\" tries to read settings from settings.json - *not implemented yet*")
	flag.StringVar(&CertFile, "cert-file", "", "relative location of cert file")
	flag.StringVar(&KeyFile, "key-file", "", "relative location of key file")
	flag.Parse()
	if len(StartDir) > 0 && strings.HasSuffix(StartDir, "/") == false {
		StartDir = StartDir + "/"
	}
}
