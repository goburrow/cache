#!/bin/sh
set -e
FILE="wiki.1191201596.gz"
curl -O "http://www.wikibench.eu/wiki/2007-10/$FILE"
gunzip "$FILE"
