#!/bin/bash
set -e

report() {
	NAME="$1"
	go test -v -run "$NAME"

	NAME=$(echo "$NAME" | tr '[:upper:]' '[:lower:]')
	./visualize-request.sh request_$NAME-*.txt
	for OUTPUT in out.*; do
		mv -v "$OUTPUT" "$NAME-requests.${OUTPUT#*.}"
	done
	./visualize-size.sh size_$NAME-*.txt
	for OUTPUT in out.*; do
		mv -v "$OUTPUT" "$NAME-cachesize.${OUTPUT#*.}"
	done
}

TRACES="Address CPP Multi2 ORMBusy Glimpse OLTP Sprite Financial WebSearch Wikipedia YouTube Zipf"
for TRACE in $TRACES; do
	report $TRACE
done
