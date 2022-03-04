package main

import "flag"

var (
	haVersion             = flag.String("ha-version", "", "Home Assistant version")
	enabledComponentsFile = flag.String("enabled-components", "", "Components to enable among the full list of integrations from Home Assistant")
	outputReqsFile        = flag.String("output-requirements", "", "Output requirements file")
	outputConstraintsFile = flag.String("output-constraints", "", "Output constraints file")
)
