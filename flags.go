package main

import "flag"

var (
	coreReqsFile         = flag.String("ha-core-requirements", "", "Home Assistant Core Requirements file")
	coreConstraintsFile  = flag.String("ha-core-constraints", "", "Home Assistant Core Constraints file")
	integrationsReqsFile = flag.String("ha-integrations", "", "Home Assistant Integrations Requirements file")
	haVersion            = flag.String("ha-version", "", "Home Assistant version")
)
