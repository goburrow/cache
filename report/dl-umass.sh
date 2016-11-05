#!/bin/sh
set -e
FILES="WebSearch1.spc.bz2 WebSearch2.spc.bz2 WebSearch3.spc.bz2"
for F in $FILES; do
	curl -O "http://skuld.cs.umass.edu/traces/storage/$F"
	bzip2 -d "$F"
done
