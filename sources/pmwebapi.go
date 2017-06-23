package sources

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type GetToken struct {
	Context int
}

type MetricHeader struct {
	Metrics []PcpMetrics
}

type PcpMetrics struct {
	Name        string
	Pmid        int64
	Indom       int64
	TextOneline string `json:"text-oneline"`
	TextHelp    string `json:"text-help"`
	Sem         string
	Units       string
	Type        string
}

type TimestampHeader struct {
	Timestamp
	Values []GetValues
}

type Timestamp struct {
	Seconds      string `json:"s"`
	Microseconds string `json:"us"`
}

type GetValues struct {
	Pmid         int64
	Name         string
	Seconds      string `json:"s"`
	Microseconds string `json:"us"`
	Instances    []InstanceList
}

type InstanceList struct {
	Instance int64
	Value    float64
}

type pcpPmwebapiMetric struct {
	Labels      []string
	Labelvalues []string
	Name        string
	TextHelp    string
	Value       float64
}

type pcpPmwebapiSource struct {
	pcpPmwebapiMetrics []pcpPmwebapiMetric
}

func init() {
	Factories["pmwebapi"] = newPcpSource
}

func gaugeMetric(labels []string, labelvalues []string, name string, textHelp string, value float64) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", name),
			textHelp,
			labels,
			nil,
		),
		prometheus.GaugeValue,
		value,
		labelvalues...,
	)
}

func (p *pcpPmwebapiSource) Update(ch chan<- prometheus.Metric) error {
	for _, metric := range p.pcpPmwebapiMetrics {
		ch <- gaugeMetric(metric.Labels, metric.Labelvalues, metric.Name, metric.TextHelp, metric.Value)
	}
	return nil
}

func getRequest(url string) []byte {
	getResponse, err := http.Get(url)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	defer getResponse.Body.Close()
	getResponseBody, err := ioutil.ReadAll(getResponse.Body)
	if err != nil {
		log.Fatal("Unable to read HTTP response:", err)
	}
	return getResponseBody
}

func newPcpSource() (PcpSource, error) {
	var p pcpPmwebapiSource
	tokenBody := getRequest("http://localhost:44323/pmapi/context?hostname=localhost")

	tokenMap := make(map[string]float64)
	err := json.Unmarshal([]byte(tokenBody), &tokenMap)

	if err != nil {
		log.Fatal("NewRequest: ", err)
	}

	contextNumber := strconv.FormatFloat(tokenMap["context"], 'f', -1, 64)
	getMetricsBody := getRequest("http://localhost:44323/pmapi/" + contextNumber + "/_metric")

	metricHeaderVar := &MetricHeader{}
	err = json.Unmarshal([]byte(getMetricsBody), metricHeaderVar)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}

	for _, pcpMetrics := range metricHeaderVar.Metrics {
		if strings.ToUpper(pcpMetrics.Type) == "STRING" {
			continue
		}
		pmidNumber := strconv.FormatInt(pcpMetrics.Pmid, 10)
		getFinalMetricsBody := getRequest("http://localhost:44323/pmapi/" + contextNumber + "/_fetch?pmids=" + pmidNumber)

		timestampHeaderVar := &TimestampHeader{}
		err = json.Unmarshal([]byte(getFinalMetricsBody), timestampHeaderVar)
		if err != nil {
			log.Fatal("NewRequest: ", err)
		}
		for _, getValues := range timestampHeaderVar.Values {
			for _, instanceList := range getValues.Instances {
				metric := pcpPmwebapiMetric{
					Labels:      []string{"instance"},
					Labelvalues: []string{strconv.FormatInt(instanceList.Instance, 10)},
					Name:        getValues.Name,
					TextHelp:    pcpMetrics.TextHelp,
					Value:       instanceList.Value,
				}
				p.pcpPmwebapiMetrics = append(p.pcpPmwebapiMetrics, metric)
			}
		}
	}
	return &p, nil
}
