.SHELL := /bin/bash

TAG:=`git describe --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)

build:
	go build -ldflags "$(LDFLAGS)" -o benchcli


