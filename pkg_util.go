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
	"sort"
	"strings"

	"github.com/tuxdude/zzzlog"
	"github.com/tuxdude/zzzlogi"
)

const (
	pkgConstraintInclude = "-c homeassistant/package_constraints.txt"
	integsHeader         = "# Home Assistant Core, full dependency set"
	integsReqsInclude    = "-r requirements.txt"
	componentPrefix      = "homeassistant.components"
	componentLinePrefix  = "# " + componentPrefix + "."
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
	if *integsReqsFile == "" {
		log.Errorf("-ha-integrations cannot be empty")
		return -1
	}
	if *haVersion == "" {
		log.Errorf("-ha-version cannot be empty")
		return -1
	}

	constraints, err := parseConstraintsOrReqsFile(*coreConstraintsFile, false)
	if err != nil {
		log.Errorf("Parsing core constraints failed, reason: %v", err)
		return -1
	}
	log.Debugf("Core Constraints: %v", constraints)
	log.Infof("Unique Core Constraints: %d", len(constraints))

	reqs, err := parseConstraintsOrReqsFile(*coreReqsFile, true)
	if err != nil {
		log.Errorf("Parsing core requirements failed, reason: %v", err)
		return -1
	}
	log.Debugf("Core Reqs: %v", reqs)
	log.Infof("Unique Core Requirements: %d", len(reqs))

	integs, err := parseIntegrationsFile(*integsReqsFile)
	if err != nil {
		log.Errorf("Parsing integrations requirements failed, reason: %v", err)
		return -1
	}
	log.Debugf("Integ Reqs: %v", integs)
	log.Infof("Unique components in Integrations: %d", len(integs))

	return 0
}

func parseConstraintsOrReqsFile(file string, firstLineWithConstraint bool) (dependencies, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseConstraintsOrReqs(f, firstLineWithConstraint)
}

func parseConstraintsOrReqs(reader io.Reader, firstLineWithConstraint bool) (dependencies, error) {
	s := bufio.NewScanner(reader)
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

	sort.Strings(deps)
	return deps, nil
}

func parseIntegrationsFile(file string) (integrations, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseIntegrations(f)
}

func parseIntegrations(reader io.Reader) (integrations, error) {
	s := bufio.NewScanner(reader)
	if !s.Scan() {
		return nil, s.Err()
	}
	if s.Text() != integsHeader {
		return nil, fmt.Errorf("First line must contain %q, found %q instead", integsHeader, s.Text())
	}

	if !s.Scan() {
		return nil, s.Err()
	}
	if s.Text() != integsReqsInclude {
		return nil, fmt.Errorf("Second line must contain %q, found %q instead", pkgConstraintInclude, s.Text())
	}

	var cmps []string
	integs := make(integrations)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" {
			if strings.HasPrefix(line, componentLinePrefix) {
				// Home Assistant Component.
				c := strings.TrimPrefix(line, componentLinePrefix)
				cmps = append(cmps, c)
			} else if strings.HasPrefix(line, "# ") {
				// Commented out dependency.
				// TODO: Make a stronger check here.
				log.Debugf("Ignoring possibly commented dependency %q", line)
				cmps = nil
			} else {
				// Dependency.
				for _, c := range cmps {
					integs[c] = append(integs[c], line)
				}
				cmps = nil
			}
		} else if cmps != nil {
			// Empty line.
			log.Fatalf("Assertion failed - cmps expected to be nil when we see an empty line, but instead contains: %v", cmps)
		}
	}

	err := s.Err()
	if err != nil {
		return nil, err
	}

	for _, deps := range integs {
		sort.Strings(deps)
	}

	return integs, nil
}

func main() {
	flag.Parse()
	os.Exit(run())
}
