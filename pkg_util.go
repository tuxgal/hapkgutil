// Command hasspkgutil parses the specified dependencies of the home assistant
// core package and the home assistant integrations, applies the specified
// allowed and denied integration component lists to generate pip compatible
// requirements and constraints files that can be used to download, install
// or even build just the wheels for these packages.
package main

import (
	"flag"
	"os"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

var (
	log = buildLogger()
)

func buildLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.MaxLevel = zzzlog.LvlWarn
	return zzzlog.NewLogger(config)
}

func run() int {
	log.Infof("Running ...")

	if *coreReqsFile == "" {
		log.Errorf("-ha-core-requirements cannot be empty")
		return -1
	}
	if *coreConstraintsFile == "" {
		log.Errorf("-ha-core-constraints cannot be empty")
		return -1
	}
	if *integrationsReqsFile == "" {
		log.Errorf("-ha-integrations cannot be empty")
		return -1
	}
	if *haVersion == "" {
		log.Errorf("-ha-version cannot be empty")
		return -1
	}

	return 0
}

func main() {
	flag.Parse()
	os.Exit(run())
}
