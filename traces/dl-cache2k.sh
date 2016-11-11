#!/bin/sh
set -e

FILES="trace-cpp.trc.bin.gz trace-glimpse.trc.bin.gz trace-mt-db-20160419-busy.trc.bin.bz2 trace-multi2.trc.bin.gz trace-oltp.trc.bin.gz trace-sprite.trc.bin.gz"
for F in $FILES; do
	if [ ! -f "$F" ]; then
		curl -L -O "https://github.com/cache2k/cache2k-benchmark/raw/master/traces/src/main/resources/org/cache2k/benchmark/traces/$F"
	fi
done
