#!/bin/sh

docker run --rm --name go-buildern -v "$PWD":/usr/src/pcp-exporter -w /usr/src/pcp-exporter go-builder