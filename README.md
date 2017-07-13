# Performance CoPilot Metrics Exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/HewlettPackard/pcp_exporter)](https://goreportcard.com/report/github.com/HewlettPackard/pcp_exporter)
[![Build Status](https://travis-ci.org/HewlettPackard/pcp_exporter.svg?branch=master)](https://travis-ci.org/HewlettPackard/pcp_exporter)

[Prometheus](https://prometheus.io/) exporter for PCP metrics.

## Getting

```
go get github.com/HewlettPackard/pcp_exporter
```

## Building


```
cd $GOPATH/src/github.com/HewlettPackard/lustre_exporter
make
```

or
```
go build github.com/HewlettPackard/pcp_exporter
```

## Running

```
./pcp_exporter
```

### What's exported?

Currently, the PCP Exporter takes PCP metrics from a locally running PMWEBD instance via the PMWEBAPI, and makes no modifications aside from assigning nonnegative instance values as labels. Our end goal is to remove the instance values as labels entirely, and use string metrics from common instances as labels instead. Also, we will augment metric names to bring them in line with Prometheus naming standards as specified here: https://prometheus.io/docs/instrumenting/writing_exporters/#naming

## Contributing

To contribute to this HPE project, you'll need to fill out a CLA (Contributor License Agreement). If you would like to contribute anything more than a bug fix (feature, architectural change, etc), please file an issue and we'll get in touch with you to have you fill out the CLA. 
