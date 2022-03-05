package main

type dependencies []string

type enabledIntegrations map[string]bool

type integrations map[string]dependencies

func (e enabledIntegrations) names() []string {
	var res []string
	for c := range e {
		res = append(res, c)
	}
	return res
}
