package pmwebapi

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
	log.Debugf("Init: pmwebapi")
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
	switch {
	case strings.HasPrefix(units, "/ "):
		unit = strings.Replace(units, "/ ", "_", 1)
	case units != "":
		unit = "_" + units
	}
	if strings.ToUpper(pcpType) == "COUNTER" {
		name += unit + "_total"
	} else {
		name += unit
	}
	return name
}

func fixNaming(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "seconds"), strings.Contains(name, "milliseconds"), strings.Contains(name, "nanoseconds"):
	case strings.Contains(name, "nanosec"):
		name = strings.Replace(name, "nanosec", "nanoseconds", -1)
	case strings.Contains(name, "millisec"):
		name = strings.Replace(name, "millisec", "milliseconds", -1)
	case strings.Contains(name, "count / sec"):
		name = strings.Replace(name, "count / sec", "count_per_second", -1)
	case strings.Contains(name, "sec"):
		name = strings.Replace(name, "sec", "seconds", -1)
	case strings.Contains(name, "mbyte"):
		name = strings.Replace(name, "mbyte", "megabytes", -1)
	case strings.Contains(name, "kbyte"):
		name = strings.Replace(name, "kbyte", "kilobytes", -1)
	case strings.Contains(name, "kilobytes"), strings.Contains(name, "megabytes"):
	case strings.Contains(name, "bytes_byte"):
		name = strings.Replace(name, "bytes_byte", "bytes", -1)
	case strings.Contains(name, "bytes"):
	case strings.Contains(name, "byte"):
		name = strings.Replace(name, "byte", "bytes", -1)
	case strings.Contains(name, "failcnt"):
		name = strings.Replace(name, "failcnt", "failcount", -1)
	case strings.Contains(name, " / "):
		name = strings.Replace(name, " / ", "_per_", -1)
	case strings.Contains(name, "/"):
		name = strings.Replace(name, "/", "_per_", -1)
	}
	switch {
	case strings.Contains(name, "seconds / count"):
		name = strings.Replace(name, "seconds / count", "seconds_per_count", -1)
	case strings.Contains(name, "mbyte / seconds"):
		name = strings.Replace(name, "mbyte / seconds", "megabytes_per_second", -1)
	case strings.Contains(name, "byte / seconds"):
		name = strings.Replace(name, "byte / seconds", "bytes_per_second", -1)
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
		pcpMetrics.Name = fixNaming(pcpMetrics.Name)
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
