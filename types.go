package main

type dependencies []string

type selectedIntegrations map[string]bool

type integrations map[string]dependencies

func (s selectedIntegrations) names() []string {
	var res []string
	for c := range s {
		res = append(res, c)
	}
	return res
}
