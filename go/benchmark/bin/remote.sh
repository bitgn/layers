#!/bin/bash
set -e

REMOTE=vault
PROJ=github.com/bitgn/layers
DEST=/tmp


if [ ! -f ~/.bitgn_tester ]
then
    echo "Need target machine"
    exit 1
fi

TARGET=$(cat ~/.bitgn_tester)

git push

ssh $REMOTE PROJ=$PROJ TARGET=$TARGET 'bash -s' <<'ENDSSH'
    set -e
    cd $GOPATH/src/$PROJ/go/benchmark
    git pull
    make build
    echo Copy to $TARGET
    scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no  benchcli $TARGET:
ENDSSH




