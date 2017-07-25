package sources

import (
	"github.com/prometheus/client_golang/prometheus"
)

//Namespace constant
const Namespace = "pcp"

//Factories variable to create a map
var Factories = make(map[string]func() (PcpSource, error))

//PcpSource interface
type PcpSource interface {
	Update(ch chan<- prometheus.Metric) (err error)
}
