// memstat collector
// this will :
//  - call mdb -k ::memstat
//  - gather memory metrics
//  - feed the collector

package collector

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"

	// Prometheus Go toolset

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// GZFreeMemCollector declares the data type within the prometheus metrics package.
type GZFreeMemCollector struct {
	gzMemPerc *prometheus.GaugeVec
	gzMemMB   *prometheus.GaugeVec
}

// NewGZFreeMemExporter returns a newly allocated exporter GZFreeMemCollector.
// It exposes the total free memory of the CN.
func NewGZFreeMemExporter() (*GZFreeMemCollector, error) {
	return &GZFreeMemCollector{
		gzMemPerc: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_percent",
			Help: "Memory percentage consumed by type.",
		}, []string{"memory"}),
		gzMemMB: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_mb",
			Help: "Memory megabytes consumed by type.",
		}, []string{"memory"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZFreeMemCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzMemPerc.Describe(ch)
	e.gzMemMB.Describe(ch)
}

// Collect fetches the stats.
func (e *GZFreeMemCollector) Collect(ch chan<- prometheus.Metric) {
	e.mdb()
	e.gzMemPerc.Collect(ch)
	e.gzMemMB.Collect(ch)
}

func (e *GZFreeMemCollector) mdb() {
	var stdout bytes.Buffer

	c := exec.Command("mdb", "-k")
	c.Stdout = &stdout
	c.Stderr = os.Stderr
	c.Stdin = bytes.NewBufferString("::memstat\n")

	if err := c.Run(); err != nil {
		log.Error(err)
	}

	err := e.parseMdbOutput(stdout.String())
	if err != nil {
		log.Errorf("error on parsing mdb: %v", err)
	}
}

func (e *GZFreeMemCollector) parseMdbOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[3 : l-1] {
		cnt := 0
		parsedLine := strings.FieldsFunc(line, func(r rune) bool {
			if r != ' ' {
				cnt = 0
				return false
			} else if r == ' ' && cnt > 0 {
				return true
			}
			cnt++

			return false
		})

		// Reached end of pages list
		if len(parsedLine) == 0 {
			return nil
		}
		memType := strings.TrimSpace(parsedLine[0])
		bytes, err := strconv.ParseFloat(strings.TrimSpace(parsedLine[2]), 64)
		if err != nil {
			return err
		}
		percentage, err := strconv.ParseFloat(strings.Replace(parsedLine[3], "%", "", -1), 64)
		if err != nil {
			return err
		}

		e.gzMemPerc.With(prometheus.Labels{"memory": memType}).Set(percentage)
		e.gzMemMB.With(prometheus.Labels{"memory": memType}).Set(bytes)
	}
	return nil
}
