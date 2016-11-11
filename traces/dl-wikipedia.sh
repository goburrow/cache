#!/bin/sh
set -e
FILES="wiki.1191201596.gz"
for F in $FILES; do
	if [ ! -f "$F" ]; then
		curl -O "http://www.wikibench.eu/wiki/2007-10/$F"
	fi
done
