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
	S  string
	Us string
}

type GetValues struct {
	Pmid      int64
	Name      string
	S         string
	Us        string
	Instances []InstanceList
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

type pcpPmwebapiSource struct {
	pcpPmwebapiMetrics []pcpPmwebapiMetric
}

func (p *pcpPmwebapiSource) Update(ch chan<- prometheus.Metric) error {
	for _, metric := range p.pcpPmwebapiMetrics {
		ch <- gaugeMetric(metric.Labels, metric.Labelvalues, metric.Name, metric.TextHelp, metric.Value)
	}
	return nil
}
func newPcpSource() (PcpSource, error) {

	var p pcpPmwebapiSource

	tokenResp, err := http.Get("http://localhost:44323/pmapi/context?hostname=localhost")
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	defer tokenResp.Body.Close()

	tokenBody, _ := ioutil.ReadAll(tokenResp.Body)
	tokenMap := make(map[string]float64)
	err = json.Unmarshal([]byte(tokenBody), &tokenMap)

	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	contextNumber := strconv.FormatFloat(tokenMap["context"], 'f', -1, 64)

	initialMetricsUrl := ("http://localhost:44323/pmapi/" + contextNumber + "/_metric")

	getMetrics, err := http.Get(initialMetricsUrl)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	defer getMetrics.Body.Close()

	getMetricsBody, _ := ioutil.ReadAll(getMetrics.Body)

	metricHeaderVar := &MetricHeader{}
	err = json.Unmarshal([]byte(getMetricsBody), metricHeaderVar)
	if err != nil {
		log.Fatal("NewRequest: ", err)
	}

	var finalMetricsUrl string

	for _, pcpMetrics := range metricHeaderVar.Metrics {
		if strings.ToUpper(pcpMetrics.Type) == "STRING" {
			continue
		}
		pmidNumber := strconv.FormatInt(pcpMetrics.Pmid, 10)
		finalMetricsUrl = ("http://localhost:44323/pmapi/" + contextNumber + "/_fetch?pmids=" + pmidNumber)

		getFinalMetrics, err := http.Get(finalMetricsUrl)
		if err != nil {
			log.Fatal("NewRequest: ", err)
		}
		defer getFinalMetrics.Body.Close()

		getFinalMetricsBody, _ := ioutil.ReadAll(getFinalMetrics.Body)

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
