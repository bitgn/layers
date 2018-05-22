#!/bin/bash
set -e
FOLDER=$1
OPEN=$2

python3 report.py -source $FOLDER

if [ ! -z "$OPEN" ]; then
    open $FOLDER/index.html
fi
