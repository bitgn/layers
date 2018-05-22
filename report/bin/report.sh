#!/bin/bash
FOLDER=$1
python3 report.py -source $FOLDER && open summary.png
