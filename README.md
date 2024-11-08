# Homelab Home Assistant Package Utility (hapkgutil)

[![Build](https://github.com/tuxgal/hapkgutil/actions/workflows/build.yml/badge.svg)](https://github.com/tuxgal/hapkgutil/actions/workflows/build.yml) [![Tests](https://github.com/tuxgal/hapkgutil/actions/workflows/tests.yml/badge.svg)](https://github.com/tuxgal/hapkgutil/actions/workflows/tests.yml) [![Lint](https://github.com/tuxgal/hapkgutil/actions/workflows/lint.yml/badge.svg)](https://github.com/tuxgal/hapkgutil/actions/workflows/lint.yml) [![CodeQL](https://github.com/tuxgal/hapkgutil/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/tuxgal/hapkgutil/actions/workflows/codeql-analysis.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/tuxgal/hapkgutil)](https://goreportcard.com/report/github.com/tuxgal/hapkgutil)

A package utility binary written in go that parses the
[`Home Assistant`](https://home-assistant.io) core package and the
list of supported integrations, filters only the integrations of interest
within tuxgal Homelab and generates the package requirements and
constraints list.

The generated package requirements and constraints list are used by the
[`tuxgalhomelab/docker-image-home-assistant`](https://github.com/tuxgalhomelab/docker-image-home-assistant)
repository's Dockerfile to build a home assistant container image that is
used in tuxgal's Homelab setup.

This tool is still generic that you can supply a Home Assistant release
version and a custom list of enabled home assistant integration components
to generate the requirements and constraints.
