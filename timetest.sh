#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "USAGE: ./timetest.sh repeat input"
    exit 1
fi

filename=`basename "$2"`.results

for i in $(eval echo {0.."$1"}); do
    for j in 1 2 4 8 16 32; do
        echo -ne "$j " >> "$filename"
        ./paratype -time -procs="$j" -infile="$2" >> "$filename"        
    done
done
