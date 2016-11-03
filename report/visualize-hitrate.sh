#!/bin/bash
FORMAT="svg"
PLOTARG=""
START=1
STEP=10
COUNT=1000

for f in "$@"; do
	if [ ! -z "$PLOTARG" ]; then
		PLOTARG="$PLOTARG,"
	fi
	NAME="$(basename "$f")"
	NAME="${NAME%.*}"
	NAME="${NAME#report-}"
	PLOTARG="$PLOTARG '$f' every $STEP::$((START))::$((START+COUNT-1)) using 1:3 with lines title '$NAME'"
done

ARG="set datafile separator ',';\
	set xlabel 'Requests';\
	set ylabel 'Hit Rate';\
	set key bottom right;\
	set terminal $FORMAT;\
	plot $PLOTARG"
if [ "$FORMAT" = "dumb" ]; then
	gnuplot -e "$ARG"
else
	gnuplot -e "$ARG" > "out.$FORMAT"
fi
