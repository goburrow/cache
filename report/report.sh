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

report Wikipedia
report YouTube
report Zipf
