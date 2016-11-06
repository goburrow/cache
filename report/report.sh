#!/bin/sh
set -e
go test -v -run Wikipedia
./visualize-hitrate.sh wikipedia-*.txt
mv out.svg wikipedia.svg

go test -v -run YouTube
./visualize-hitrate.sh youtube-*.txt
mv out.svg youtube.svg

go test -v -run Zipf
./visualize-hitrate.sh zipf-*.txt
mv out.svg zipf.svg
