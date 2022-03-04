# Home Lab Home Assistant Package Utility (hasspkgutil)

[![Build](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/build.yml/badge.svg)](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/build.yml) [![Tests](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/tests.yml/badge.svg)](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/tests.yml) [![Lint](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/lint.yml/badge.svg)](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/lint.yml) [![CodeQL](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/TuxdudeHomeLab/hasspkgutil/actions/workflows/codeql-analysis.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/tuxdudehomelab/hasspkgutil)](https://goreportcard.com/report/github.com/tuxdudehomelab/hasspkgutil)

A package utility binary written in go that parses the
[`Home Assistant`](https://home-assistant.io) core package and the
list of supported integrations, filters only the integrations of interest
within Tuxdude Home Lab and generates the package requirements and
constraints list.

The generated package requirements and constraints list are used by the
[`TuxdudeHomeLab/docker-image-home-assistant-wheels`](https://github.com/TuxdudeHomeLab/docker-image-home-assistant-wheels)
repository Dockerfile to build a container image with these wheels, and
then used by the
[`TuxdudeHomeLab/docker-image-home-assistant`](https://github.com/TuxdudeHomeLab/docker-image-home-assistant)
repository's Dockerfile to build a home assistant container image that is
used in Tuxdude's Home Lab setup.

This tool is still generic that you can supply a Home Assistant release
version and a custom list of enabled home assistant integration components
to generate the requirements and constraints.
