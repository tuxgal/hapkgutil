// Command hapkgutil parses the specified dependencies of the home assistant
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
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tuxgal/tuxlog"
	"github.com/tuxgal/tuxlogi"
)

const (
	pkgConstraintInclude  = "-c homeassistant/package_constraints.txt"
	integsReqsInclude     = "-r requirements.txt"
	componentPrefix       = "components."
	integPrefix           = "# homeassistant."
	coreConstraintsURLFmt = "http://raw.githubusercontent.com/home-assistant/core/%s/homeassistant/package_constraints.txt"
	coreReqsURLFmt        = "http://raw.githubusercontent.com/home-assistant/core/%s/requirements.txt"
	integsURLFmt          = "http://raw.githubusercontent.com/home-assistant/core/%s/requirements_all.txt"
)

var (
	log = buildLogger()
)

func buildLogger() tuxlogi.Logger {
	config := tuxlog.NewConsoleLoggerConfig()
	config.MaxLevel = tuxlog.LvlInfo
	return tuxlog.NewLogger(config)
}

func run() int {
	if *generateMode && *updateMode {
		log.Errorf("Both -mode-generate and -mode-update cannot be set")
		return -1
	}
	if !*generateMode && !*updateMode {
		log.Errorf("At least one of -mode-generate or -mode-update must be set")
		return -1
	}
	var mode execMode
	if *generateMode {
		mode = modeGenerate
	} else if *updateMode {
		mode = modeUpdate
	}

	if *haVersion == "" {
		log.Errorf("-ha-version cannot be empty")
		return -1
	}
	if *enabledIntegsFile == "" {
		log.Errorf("-enabled-integrations cannot be empty")
		return -1
	}
	if *disabledIntegsFile == "" {
		log.Errorf("-disabled-integrations cannot be empty")
		return -1
	}
	if mode == modeGenerate {
		if *outputReqsFile == "" {
			log.Errorf("-output-requirements cannot be empty in generate mode")
			return -1
		}
		if *outputConstraintsFile == "" {
			log.Errorf("-output-constraints cannot be empty in generate mode")
			return -1
		}
	}

	coreConstraintsURL := fmt.Sprintf(coreConstraintsURLFmt, *haVersion)
	coreReqsURL := fmt.Sprintf(coreReqsURLFmt, *haVersion)
	integsURL := fmt.Sprintf(integsURLFmt, *haVersion)

	constraints, err := parseConstraintsOrReqsFile(coreConstraintsURL, false)
	if err != nil {
		log.Errorf("Parsing core constraints failed, reason: %v", err)
		return -1
	}
	log.Debugf("Core Constraints: %v", constraints)
	log.Infof("Unique Core Constraints: %d", len(constraints))

	reqs, err := parseConstraintsOrReqsFile(coreReqsURL, true)
	if err != nil {
		log.Errorf("Parsing core requirements failed, reason: %v", err)
		return -1
	}
	log.Debugf("Core Reqs: %v", reqs)
	log.Infof("Unique Core Requirements: %d", len(reqs))

	integs, err := parseIntegrationsFile(integsURL)
	if err != nil {
		log.Errorf("Parsing integrations requirements failed, reason: %v", err)
		return -1
	}
	log.Debugf("Integ Reqs: %v", integs)
	log.Infof("Unique Integrations: %d", len(integs))

	if mode == modeUpdate {
		log.Infof("\n\n")
		log.Infof("Before updating enabled/disabled integrations")
	}

	enabledIntegs, err := parseSelectedIntegsFile(*enabledIntegsFile)
	if err != nil {
		log.Errorf("Parsing enabled integrations failed, reason: %v", err)
		return -1
	}
	log.Debugf("Enabled integrations: %v", enabledIntegs)
	log.Infof("Unique enabled integrations: %d", len(enabledIntegs))

	disabledIntegs, err := parseSelectedIntegsFile(*disabledIntegsFile)
	if err != nil {
		log.Errorf("Parsing disabled integrations failed, reason: %v", err)
		return -1
	}
	log.Debugf("Disabled integrations: %v", disabledIntegs)
	log.Infof("Unique disabled integrations: %d", len(disabledIntegs))

	switch mode {
	case modeGenerate:
		err = validateSelectedIntegs(integs, enabledIntegs, disabledIntegs)
		if err != nil {
			log.Errorf("Enabled/Disabled integrations validation failed, reason: %v", err)
			return -1
		}

		err = writeConstraintsFile(constraints)
		if err != nil {
			log.Errorf("Failed to output constraints, reason: %v", err)
			return -1
		}

		err = writeReqsFile(reqs, integs, enabledIntegs)
		if err != nil {
			log.Errorf("Failed to output requirements, reason: %v", err)
			return -1
		}
	case modeUpdate:
		err = updateSelectedIntegs(integs, enabledIntegs, disabledIntegs)
		if err != nil {
			log.Errorf("Unable to update Enabled/Disabled integrations, reason: %v", err)
			return -1
		}

		// Validate the updated set of enabled and disabled integrations.
		err = validateSelectedIntegs(integs, enabledIntegs, disabledIntegs)
		if err != nil {
			log.Errorf("Enabled/Disabled integrations validation after update failed, reason: %v", err)
			return -1
		}

		log.Infof("\n\n")
		log.Infof("After updating enabled/disabled integrations")
		log.Debugf("Enabled integrations: %v", enabledIntegs)
		log.Infof("Unique enabled integrations: %d", len(enabledIntegs))
		log.Debugf("Disabled integrations: %v", disabledIntegs)
		log.Infof("Unique disabled integrations: %d", len(disabledIntegs))

		err = writeSelectedIntegsFile(enabledIntegs, *enabledIntegsFile)
		if err != nil {
			log.Errorf("Failed to write updated enabled integrations file, reason: %v", err)
			return -1
		}

		err = writeSelectedIntegsFile(disabledIntegs, *disabledIntegsFile)
		if err != nil {
			log.Errorf("Failed to write updated disabled integrations file, reason: %v", err)
			return -1
		}
	default:
		log.Errorf("Something went wrong, since both generate mode and update mode flags are false")
		return -1
	}

	return 0
}

func parseConstraintsOrReqsFile(url string, firstLineWithConstraint bool) (dependencies, error) {
	f, err := downloadFile(url)
	if err != nil {
		return nil, err
	}

	return parseConstraintsOrReqs(strings.NewReader(f), firstLineWithConstraint)
}

func parseConstraintsOrReqs(reader io.Reader, firstLineWithConstraint bool) (dependencies, error) {
	s := bufio.NewScanner(reader)
	line := ""
	for {
		if !s.Scan() {
			return nil, s.Err()
		}
		if l, ok := parseLine(s); ok {
			line = l
			break
		}
	}
	if firstLineWithConstraint && line != pkgConstraintInclude {
		return nil, fmt.Errorf("first line must contain %q, found %q instead", pkgConstraintInclude, line)
	}

	var deps dependencies
	for s.Scan() {
		if l, ok := parseLine(s); ok {
			deps = append(deps, l)
		}
	}

	err := s.Err()
	if err != nil {
		return nil, err
	}

	sort.Strings(deps)
	return deps, nil
}

func parseIntegrationsFile(url string) (integrations, error) {
	f, err := downloadFile(url)
	if err != nil {
		return nil, err
	}

	return parseIntegrations(strings.NewReader(f))
}

func parseIntegrations(reader io.Reader) (integrations, error) {
	s := bufio.NewScanner(reader)
	line := ""
	for {
		if !s.Scan() {
			return nil, s.Err()
		}
		if l, ok := parseLine(s); ok {
			line = l
			break
		}
	}
	if line != integsReqsInclude {
		return nil, fmt.Errorf("beginning of the file must contain %q, found %q instead", integsReqsInclude, line)
	}

	var integNames []string
	integs := make(integrations)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" {
			if strings.HasPrefix(line, integPrefix) {
				// Home Assistant Integration.
				c := strings.TrimPrefix(line, integPrefix)
				integNames = append(integNames, c)
			} else {
				// Handle commented out dependency.
				line = strings.TrimPrefix(line, "# ")
				// Add the dependency against all integrations we read
				// earlier since the previous dependency.
				for _, c := range integNames {
					integs[c] = append(integs[c], line)
				}
				integNames = nil
			}
		} else if integNames != nil {
			// Empty line.
			log.Fatalf("Assertion failed - integration names expected to be nil when we see an empty line, but instead contains: %v", integNames)
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

func parseLine(s *bufio.Scanner) (string, bool) {
	line := strings.TrimSpace(s.Text())
	if line != "" && !strings.HasPrefix(line, "#") {
		return line, true
	}
	return "", false
}

func validateSelectedIntegs(integs integrations, enabledIntegs selectedIntegrations, disabledIntegs selectedIntegrations) error {
	var invalidEnabledIntegs []string
	for e := range enabledIntegs {
		if _, ok := integs[e]; !ok {
			invalidEnabledIntegs = append(invalidEnabledIntegs, e)
		}
	}
	if len(invalidEnabledIntegs) > 0 {
		sort.Strings(invalidEnabledIntegs)
		return fmt.Errorf("cannot find the specified enabled integrations %v in the full list of integrations", invalidEnabledIntegs)
	}

	if len(disabledIntegs) == 0 {
		return nil
	}

	var invalidDisabledIntegs []string
	var commonIntegs []string
	fullSel := make(selectedIntegrations)
	for e := range enabledIntegs {
		fullSel[e] = true
	}
	for d := range disabledIntegs {
		if fullSel[d] {
			commonIntegs = append(commonIntegs, d)
		}
		if _, ok := integs[d]; !ok {
			invalidDisabledIntegs = append(invalidDisabledIntegs, d)
		}
		fullSel[d] = true
	}

	var notFoundInSel []string
	for i := range integs {
		if !fullSel[i] {
			notFoundInSel = append(notFoundInSel, i)
		}
	}

	// Verify every specified disabled integration is valid.
	if len(invalidDisabledIntegs) > 0 {
		sort.Strings(invalidDisabledIntegs)
		return fmt.Errorf("cannot find the specified disabled integrations %v in the full list of integrations", invalidDisabledIntegs)
	}

	// Verify there is no overlap between enabled and disabled integrations lists.
	if len(commonIntegs) > 0 {
		sort.Strings(commonIntegs)
		return fmt.Errorf("integrations %v are specified in both enabled and disabled integrations lists", commonIntegs)
	}

	// Verify that the union of enabled and disabled integrations lists is the
	// same as the full integrations list we built from the Home Assistant
	// release requirements file.
	if len(notFoundInSel) > 0 {
		sort.Strings(notFoundInSel)
		return fmt.Errorf("integrations %v found in the full integrations list are not part of either enabled or disabled integrations lists specified", notFoundInSel)
	}

	return nil
}

func updateSelectedIntegs(integs integrations, enabledIntegs selectedIntegrations, disabledIntegs selectedIntegrations) error {
	var invalidEnabledIntegs []string
	for e := range enabledIntegs {
		if _, ok := integs[e]; !ok {
			invalidEnabledIntegs = append(invalidEnabledIntegs, e)
		}
	}
	sort.Strings(invalidEnabledIntegs)

	var invalidDisabledIntegs []string
	var commonIntegs []string
	var notFoundInSel []string

	if len(disabledIntegs) != 0 {
		fullSel := make(selectedIntegrations)
		for e := range enabledIntegs {
			fullSel[e] = true
		}
		for d := range disabledIntegs {
			if fullSel[d] {
				commonIntegs = append(commonIntegs, d)
			}
			if _, ok := integs[d]; !ok {
				invalidDisabledIntegs = append(invalidDisabledIntegs, d)
			}
			fullSel[d] = true
		}

		for i := range integs {
			if !fullSel[i] {
				notFoundInSel = append(notFoundInSel, i)
			}
		}

		// Verify there is no overlap between enabled and disabled integrations lists.
		if len(commonIntegs) > 0 {
			sort.Strings(commonIntegs)
			return fmt.Errorf("integrations %v are specified in both enabled and disabled integrations lists", commonIntegs)
		}
	}

	sort.Strings(invalidDisabledIntegs)
	sort.Strings(commonIntegs)
	sort.Strings(notFoundInSel)

	// Remove invalidEnabledIntegs from enabledIntegs
	for _, i := range invalidEnabledIntegs {
		delete(enabledIntegs, i)
	}
	// Remove invalidDisabledIntegs from disabledIntegs
	for _, i := range invalidDisabledIntegs {
		delete(disabledIntegs, i)
	}
	// Add notFoundInSel to disabledIntegs
	for _, i := range notFoundInSel {
		disabledIntegs[i] = true
	}

	log.Infof("UPDATE: Removed from enabled integrations: %v", invalidEnabledIntegs)
	log.Infof("UPDATE: Removed from disabled integrations: %v", invalidDisabledIntegs)
	log.Infof("UPDATE: Added to disabled integrations: %v", notFoundInSel)

	return nil
}

func downloadFile(url string) (string, error) {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := c.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download from URL: %q, reason: %v", url, err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the body while downloading from URL: %q, reason: %v", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed while downloading from URL: %q, status code: %d, body: %s", url, resp.StatusCode, string(body))
	}

	return string(body), nil
}

func parseSelectedIntegsFile(file string) (selectedIntegrations, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer f.Close()

	return parseSelectedIntegs(f)
}

func parseSelectedIntegs(reader io.Reader) (selectedIntegrations, error) {
	s := bufio.NewScanner(reader)

	integs := make(selectedIntegrations)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line != "" && !strings.HasPrefix(line, "# ") {
			if strings.Contains(line, ".") {
				integs[line] = true
			} else {
				integs[componentPrefix+line] = true
			}
		}
	}

	err := s.Err()
	if err != nil {
		return nil, err
	}

	return integs, nil
}

func writeReqsFile(reqs dependencies, integs integrations, enabledIntegs selectedIntegrations) error {
	f, err := os.OpenFile(*outputReqsFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output requirements file %q", *outputReqsFile)
	}
	//nolint:errcheck
	defer f.Close()

	return outputReqs(f, reqs, integs, enabledIntegs)
}

func writeConstraintsFile(constraints dependencies) error {
	f, err := os.OpenFile(*outputConstraintsFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output constraints file %q", *outputReqsFile)
	}
	//nolint:errcheck
	defer f.Close()

	return outputConstraints(f, constraints)
}

func writeSelectedIntegsFile(integs selectedIntegrations, outputPath string) error {
	f, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output file %q for storing the updated integrations", outputPath)
	}
	//nolint:errcheck
	defer f.Close()

	integNames := make([]string, 0)
	for i := range integs {
		integNames = append(integNames, strings.TrimPrefix(i, componentPrefix))
	}
	sort.Strings(integNames)

	count, err := outputRows(f, integNames, "integration")
	if err != nil {
		return err
	}

	log.Infof("Updated integrations file %q, wrote %d bytes", outputPath, count)
	return nil
}

func outputReqs(writer io.Writer, reqs dependencies, integs integrations, enabledIntegs selectedIntegrations) error {
	haDep := fmt.Sprintf("homeassistant==%s", *haVersion)
	deps := append(reqs, haDep)
	for c, d := range integs {
		if enabledIntegs[c] {
			deps = append(deps, d...)
		}
	}
	sort.Strings(deps)

	count, err := outputRows(writer, deps, "dependency")
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

	count, err := outputRows(writer, constraints, "dependency")
	if err != nil {
		return err
	}

	log.Infof("Output Constraints, wrote %d bytes", count)
	return nil
}

func outputRows(writer io.Writer, rows []string, rowType string) (int, error) {
	count := 0
	w := bufio.NewWriter(writer)
	//nolint:errcheck
	defer w.Flush()

	for _, dep := range rows {
		n, err := w.WriteString(dep)
		if err != nil {
			return 0, fmt.Errorf("failed to write %s to file, reason: %v", rowType, err)
		}
		count += n
		n, err = w.WriteString("\n")
		if err != nil {
			return 0, fmt.Errorf("failed to write %s to file, reason: %v", rowType, err)
		}
		count += n
	}
	return count, nil
}

func main() {
	flag.Parse()
	os.Exit(run())
}
