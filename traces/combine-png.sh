#!/bin/sh
set -e
NAMES="address wikipedia youtube zipf"
FORMAT="png"
FILES=""
for N in $NAMES; do
	FILES="$FILES $N-requests.$FORMAT $N-cachesize.$FORMAT"
done
gm montage -mode concatenate -tile 4x $FILES "report.$FORMAT"
