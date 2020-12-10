#!/bin/sh

echo "performing go get -d -v"
go get -d -v 

echo
echo "Building..."
go build -v
# make