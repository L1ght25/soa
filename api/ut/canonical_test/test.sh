#!/bin/bash

python3 test.py > data.txt

if cmp -s data.txt canondata.txt; then
    echo "OK"
else
    echo "ERROR: Data doesn't match canondata.txt"
fi
