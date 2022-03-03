package main

import "flag"

var (
	coreReqsFile          = flag.String("ha-core-requirements", "", "Home Assistant Core Requirements file")
	coreConstraintsFile   = flag.String("ha-core-constraints", "", "Home Assistant Core Constraints file")
	integsReqsFile        = flag.String("ha-integrations", "", "Home Assistant Integrations Requirements file")
	haVersion             = flag.String("ha-version", "", "Home Assistant version")
	enabledComponentsFile = flag.String("enabled-components", "", "Components to enable among the full list of integrations from Home Assistant")
	outputReqsFile        = flag.String("output-requirements", "", "Output requirements file")
	outputConstraintsFile = flag.String("output-constraints", "", "Output constraints file")
)
