package main

import "flag"

var (
	haVersion             = flag.String("ha-version", "", "Home Assistant version")
	enabledIntegsFile     = flag.String("enabled-integrations", "", "Integrations to enable among the full list of integrations from Home Assistant")
	disabledIntegsFile    = flag.String("disabled-integrations", "", "Integrations to disable among the full list of integrations from Home Assistant (used as a sanity check for ensuring full integration coverage)")
	outputReqsFile        = flag.String("output-requirements", "", "Output requirements file")
	outputConstraintsFile = flag.String("output-constraints", "", "Output constraints file")
)
