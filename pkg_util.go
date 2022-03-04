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
	if *enabledComponentsFile == "" {
		log.Errorf("-enabled-components cannot be empty")
		return -1
	}
	if *outputReqsFile == "" {
		log.Errorf("-output-requirements cannot be empty")
		return -1
	}
	if *outputConstraintsFile == "" {
		log.Errorf("-output-constraints cannot be empty")
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

	enabledCmps, err := parseComponentsFile(*enabledComponentsFile)
	if err != nil {
		log.Errorf("Parsing enabled components failed, reason: %v", err)
		return -1
	}
	log.Debugf("Enabled components: %v", enabledCmps)
	log.Infof("Unique enabled components: %d", len(enabledCmps))

	err = writeConstraintsFile(constraints)
	if err != nil {
		log.Errorf("Failed to output constraints, reason: %v", err)
		return -1
	}

	err = writeReqsFile(reqs, integs, enabledCmps)
	if err != nil {
		log.Errorf("Failed to output requirements, reason: %v", err)
		return -1
	}

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
			} else {
				// Handle commented out dependency.
				line = strings.TrimPrefix(line, "# ")
				// Add the dependency against all the components.
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

func parseComponentsFile(file string) (components, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseComponents(f)
}

func parseComponents(reader io.Reader) (components, error) {
	s := bufio.NewScanner(reader)

	cmps := make(components)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" && !strings.HasPrefix(line, "# ") {
			cmps[line] = true
		}
	}

	err := s.Err()
	if err != nil {
		return nil, err
	}

	return cmps, nil
}

func writeReqsFile(reqs dependencies, integs integrations, enabledCmps components) error {
	f, err := os.OpenFile(*outputReqsFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output requirements file %q", *outputReqsFile)
	}
	defer f.Close()

	return outputReqs(f, reqs, integs, enabledCmps)
}

func writeConstraintsFile(constraints dependencies) error {
	f, err := os.OpenFile(*outputConstraintsFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output constraints file %q", *outputReqsFile)
	}
	defer f.Close()

	return outputConstraints(f, constraints)
}

func outputReqs(writer io.Writer, reqs dependencies, integs integrations, enabledCmps components) error {
	haDep := fmt.Sprintf("homeassistant==%s", *haVersion)
	deps := append(reqs, haDep)
	for c, d := range integs {
		if enabledCmps[c] {
			deps = append(deps, d...)
		}
	}
	sort.Strings(deps)

	count, err := outputDeps(writer, deps)
	if err != nil {
		return err
	}
	log.Infof("Output Requirements, wrote %d bytes", count)
	return nil
}

func outputConstraints(writer io.Writer, constraints dependencies) error {
	haDep := fmt.Sprintf("homeassistant==%s", *haVersion)
	deps := append(constraints, haDep)
	sort.Strings(deps)

	count, err := outputDeps(writer, constraints)
	if err != nil {
		return err
	}

	log.Infof("Output Constraints, wrote %d bytes", count)
	return nil
}

func outputDeps(writer io.Writer, deps dependencies) (int, error) {
	count := 0
	w := bufio.NewWriter(writer)
	defer w.Flush()

	for _, dep := range deps {
		n, err := w.WriteString(dep)
		if err != nil {
			return 0, fmt.Errorf("failed to write dependency to file, reason: %v", err)
		}
		count += n
		n, err = w.WriteString("\n")
		if err != nil {
			return 0, fmt.Errorf("failed to write dependency to file, reason: %v", err)
		}
		count += n
	}
	return count, nil
}

func main() {
	flag.Parse()
	os.Exit(run())
}
