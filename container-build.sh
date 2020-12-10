#!/bin/sh

rm -f pcp-exporter

echo "performing go get -d -v"
go get -d -v 

echo
echo "Building..."
go build -v
# make