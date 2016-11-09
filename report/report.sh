#!/bin/sh
set -e

report() {
	NAME="$1"
	go test -v -run "$NAME"

	NAME=$(echo "$NAME" | tr '[:upper:]' '[:lower:]')
	./visualize-request.sh request_$NAME-*.txt
	mv -v out.svg $NAME-requests.svg
	./visualize-size.sh size_$NAME-*.txt
	mv -v out.svg $NAME-cachesize.svg
}

report Wikipedia
report YouTube
report Zipf
