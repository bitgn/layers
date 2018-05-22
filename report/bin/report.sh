#!/bin/bash
set -e
FOLDER=$1
python3 report.py -source $FOLDER
