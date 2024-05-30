package main

type dependencies []string

type selectedIntegrations map[string]bool

type integrations map[string]dependencies

type execMode uint8

const (
	modeGenerate execMode = iota
	modeUpdate
)
