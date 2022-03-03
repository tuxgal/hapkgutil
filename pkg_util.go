// Command hasspkgutil parses the specified dependencies of the home assistant
// core package and the home assistant integrations, applies the specified
// allowed and denied integration component lists to generate pip compatible
// requirements and constraints files that can be used to download, install
// or even build just the wheels for these packages.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

const (
	pkgConstraintInclude = "-c homeassistant/package_constraints.txt"
)

var (
	log = buildLogger()
)

func buildLogger() zzzlogi.Logger {
	config := zzzlog.NewConsoleLoggerConfig()
	config.MaxLevel = zzzlog.LvlInfo
	return zzzlog.NewLogger(config)
}

func run() int {
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

	err := parseCoreReqsFile()
	if err != nil {
		log.Errorf("Parsing core requirements failed, reason: %v", err)
		return -1
	}

	return 0
}

func parseCoreReqsFile() error {
	f, err := os.Open(*coreReqsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	return parseCoreReqs(f)
}

func parseCoreReqs(file io.Reader) error {
	s := bufio.NewScanner(file)
	if !s.Scan() {
		return s.Err()
	}
	if s.Text() != pkgConstraintInclude {
		return fmt.Errorf("First line must contain %q, found %q instead", pkgConstraintInclude, s.Text())
	}
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			log.Infof(line)
		}
	}

	return s.Err()
}

func main() {
	flag.Parse()
	os.Exit(run())
}
