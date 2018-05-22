#!/bin/bash
set -e

REMOTE=vault
PROJ=github.com/bitgn/layers
DEST=/tmp

git push

ssh $REMOTE PROJ=$PROJ DEST=$DEST 'bash -s' <<'ENDSSH'
	set -e
	cd $GOPATH/src/$PROJ/go/benchmark
	git pull
	make build
	mv benchcli $DEST
ENDSSH

scp $REMOTE:$DEST/benchcli .
