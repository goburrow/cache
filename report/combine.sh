#!/bin/sh
set -e

if [ -z "$FORMAT" ]; then
	FORMAT="png"
else
	FORMAT="${FORMAT%% *}"
fi

FILES=$(ls *-requests.$FORMAT *-cachesize.$FORMAT | sort)
gm montage -mode concatenate -tile 4x $FILES "report.$FORMAT"
