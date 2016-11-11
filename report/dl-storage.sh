#!/bin/sh
set -e
FILES="WebSearch1.spc.bz2 Financial2.spc.bz2"
for F in $FILES; do
	if [ ! -f "$F" ]; then
		curl -O "http://skuld.cs.umass.edu/traces/storage/$F"
	fi
done
