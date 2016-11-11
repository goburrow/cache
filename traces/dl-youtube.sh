#!/bin/sh
set -e
FILE="youtube_traces.tgz"
if [ ! -f "$FILE" ]; then
	curl -O "http://skuld.cs.umass.edu/traces/network/$FILE"
fi
tar xzf "$FILE"

rm youtube.parsed.*.24.dat
rm youtube.parsed.*.S1.dat

for FILE in youtube.parsed.*.dat; do
	# YYMMDD
	NAME="$(echo "$FILE" | sed -e 's/\([0-9]\{2\}\)\([0-9]\{2\}\)\([0-9]\{2\}\)\(\.dat\)/\3\1\2\4/')"
	mv "$FILE" "$NAME"
done
