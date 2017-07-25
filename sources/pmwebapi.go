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

type metricHeader struct {
	Metrics []pcpMetrics
}

type pcpMetrics struct {
	Name        string
	Pmid        int64
	Indom       int64
	TextOneline string `json:"text-oneline"`
	TextHelp    string `json:"text-help"`
	Sem         string
	Units       string
	Type        string
}

type timestampHeader struct {
	Values []getValues
}

type getValues struct {
	Pmid         int64
	Name         string
	Seconds      string `json:"s"`
	Microseconds string `json:"us"`
	Instances    []instanceList
}

type instanceList struct {
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

func unmarshal(text []byte, t interface{}) {
	err := json.Unmarshal(text, t)
	if err != nil {
		log.Fatal("New Request: ", err)
	}
}

func typeLabel(units string, pcpType string, name string) string {
	unit := ""
	if units != "" {
		unit = "_" + units
	}
	if strings.ToUpper(pcpType) == "COUNTER" {
		name += unit + "_total"
	} else {
		name += unit
	}
	return strings.Replace(name, ".", "_", -1)
}

func newPcpSource() (PcpSource, error) {
	var p pcpPmwebapiSource
	tokenBody := getRequest("http://localhost:44323/pmapi/context?hostname=localhost")

	tokenMap := make(map[string]float64)
	unmarshal([]byte(tokenBody), &tokenMap)

	contextNumber := strconv.FormatFloat(tokenMap["context"], 'f', -1, 64)
	getMetricsBody := getRequest("http://localhost:44323/pmapi/" + contextNumber + "/_metric")

	metricHeaderVar := &metricHeader{}
	unmarshal([]byte(getMetricsBody), metricHeaderVar)

	for _, pcpMetrics := range metricHeaderVar.Metrics {
		if strings.ToUpper(pcpMetrics.Type) == "STRING" {
			continue
		}

		pmidNumber := strconv.FormatInt(pcpMetrics.Pmid, 10)
		getFinalMetricsBody := getRequest("http://localhost:44323/pmapi/" + contextNumber + "/_fetch?pmids=" + pmidNumber)

		timestampHeaderVar := &timestampHeader{}
		unmarshal([]byte(getFinalMetricsBody), timestampHeaderVar)

		pcpMetrics.Name = typeLabel(pcpMetrics.Units, pcpMetrics.Type, pcpMetrics.Name)

		for _, getValues := range timestampHeaderVar.Values {
			for _, instanceList := range getValues.Instances {
				labelNew := []string{}
				labelValuesNew := []string{}
				if instanceList.Instance != -1 {
					labelNew = append(labelNew, "instance")
					labelValuesNew = append(labelValuesNew, strconv.FormatInt(instanceList.Instance, 10))
				}
				metric := pcpPmwebapiMetric{
					Labels:      labelNew,
					Labelvalues: labelValuesNew,
					Name:        pcpMetrics.Name,
					TextHelp:    pcpMetrics.TextHelp,
					Value:       instanceList.Value,
				}
				p.pcpPmwebapiMetrics = append(p.pcpPmwebapiMetrics, metric)
			}
		}
	}
	return &p, nil
}
