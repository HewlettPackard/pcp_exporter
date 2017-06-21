package sources

import (
	"github.com/prometheus/client_golang/prometheus"
)

const Namespace = "pcp"

var Factories = make(map[string]func() (PcpSource, error))

type PcpSource interface {
	Update(ch chan<- prometheus.Metric) (err error)
}
