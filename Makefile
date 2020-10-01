SHELL := /usr/bin/env bash -o pipefail
GOPKG ?= github.com/MrEhbr/homekit
DOCKER_IMAGE ?=	MrEhbr/homekit
GOBINS ?= .

include rules.mk
