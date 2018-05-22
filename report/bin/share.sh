#!/bin/bash
set -e
FOLDER=$1
if [ ! -n "$FOLDER" ]
then
    echo "Need target folder"
    exit 1
fi

aws s3 sync $FOLDER s3://r.bitgn.com/$(basename $FOLDER) --profile bitgn-deploy --cache-control "public, max-age=3600" --region eu-west-1
open https://r.bitgn.com/$(basename $FOLDER)
