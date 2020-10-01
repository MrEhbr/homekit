SHELL := /usr/bin/env bash -o pipefail
GOPKG ?= github.com/MrEhbr/homekit
DOCKER_IMAGE ?=	MrEhbr/homekit
GOBINS ?= .
PROTO_PATH := .
PROTOC_GEN_GO_OUT := .
PROTOC_GEN_GO_OPT := plugins=grpc

include rules.mk
