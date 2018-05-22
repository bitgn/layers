.SHELL := /bin/bash

TAG:=`git describe --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)

build:
	echo "Flags are $(LDFLAGS)"
	go build -ldflags "$(LDFLAGS)" -o benchcli


