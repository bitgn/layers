#!/bin/bash
set -e

if [ ! -f ~/.bitgn_tester ]
then
    echo "Need target machine"
    exit 1
fi

rm -rf /tmp/bench-import
mkdir -p /tmp/bench-import

scp -r $(cat ~/.bitgn_tester):bench-\* /tmp/bench-import

for d in /tmp/bench-import/* ; do
    echo "$d"
    bin/report.sh $d
done
