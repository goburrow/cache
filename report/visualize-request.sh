#!/bin/bash
FORMAT='svg size 400,300 font "Helvetica,10"'
PLOTARG=""

for f in "$@"; do
	if [ ! -z "$PLOTARG" ]; then
		PLOTARG="$PLOTARG,"
	fi
	NAME="$(basename "$f")"
	NAME="${NAME%.*}"
	NAME="${NAME#*_}"
	PLOTARG="$PLOTARG '$f' every ::1 using 1:3 with lines title '$NAME'"
done

ARG="set datafile separator ',';\
	set xlabel 'Requests';\
	set xtics rotate by -45 offset -1;\
	set ylabel 'Hit Rate';\
	set yrange [0:];\
	set key bottom right;\
	set colors classic;\
	set terminal $FORMAT;\
	plot $PLOTARG"
if [ "$FORMAT" = "dumb" ]; then
	gnuplot -e "$ARG"
else
	gnuplot -e "$ARG" > "out.${FORMAT%% *}"
fi
