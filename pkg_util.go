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

	deps, err := parseCoreReqsFile()
	if err != nil {
		log.Errorf("Parsing core requirements failed, reason: %v", err)
		return -1
	}
	log.Infof("Core Reqs: %v", deps)

	return 0
}

func parseCoreReqsFile() (dependencies, error) {
	f, err := os.Open(*coreReqsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseCoreReqs(f)
}

func parseCoreReqs(file io.Reader) (dependencies, error) {
	return parseReqsOrConstraints(file, true)
}

func parseReqsOrConstraints(file io.Reader, firstLineWithConstraint bool) (dependencies, error) {
	s := bufio.NewScanner(file)
	if !s.Scan() {
		return nil, s.Err()
	}
	if firstLineWithConstraint && s.Text() != pkgConstraintInclude {
		return nil, fmt.Errorf("First line must contain %q, found %q instead", pkgConstraintInclude, s.Text())
	}

	var deps dependencies
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			deps = append(deps, line)
		}
	}

	err := s.Err()
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func main() {
	flag.Parse()
	os.Exit(run())
}
