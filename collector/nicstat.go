// nicstat collector
// this will :
//  - call nicstat
//  - gather network metrics
//  - feed the collector

package collector

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"

	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// var devices = []string{"ixgbe0", "ixgbe1", "loop0"}

// GZNICUsageCollector declares the data type within the prometheus metrics
// package.
type GZNICUsageCollector struct {
	devices         []string
	gzNICUsageRead  *prometheus.GaugeVec
	gzNICUsageWrite *prometheus.GaugeVec
}

// NewGZNICUsageExporter returns a newly allocated exporter GZNICUsageCollector.
// It exposes the network bandwidth used by the MLAG interface
func NewGZNICUsageExporter(nics ...string) (*GZNICUsageCollector, error) {
	return &GZNICUsageCollector{
		devices: nics,
		gzNICUsageRead: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_receive_kilobytes",
			Help: "NIC receive statistic in KBytes.",
		}, []string{"device"}),
		gzNICUsageWrite: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_transmit_kilobytes",
			Help: "NIC transmit statistic in KBytes.",
		}, []string{"device"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZNICUsageCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzNICUsageRead.Describe(ch)
	e.gzNICUsageWrite.Describe(ch)
}

// Collect fetches the stats.
func (e *GZNICUsageCollector) Collect(ch chan<- prometheus.Metric) {
	e.nicstat()
	e.gzNICUsageRead.Collect(ch)
	e.gzNICUsageWrite.Collect(ch)
}

func (e *GZNICUsageCollector) nicstat() {
	var wg sync.WaitGroup
	wg.Add(len(e.devices))
	for _, device := range e.devices {
		// Do these in parallel since it takes 2 seconds for each interface
		go func(device string) {
			defer wg.Done()
			out, eerr := exec.Command("nicstat", "-i", device, "1", "2").Output()
			if eerr != nil {
				log.Errorf("error on executing nicstat: %v", eerr)
				return
			}
			perr := e.parseNicstatOutput(string(out), device)
			if perr != nil {
				log.Errorf("error on parsing nicstat: %v", perr)
			}
		}(device)
	}
	wg.Wait()
}

func (e *GZNICUsageCollector) parseNicstatOutput(out, device string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[2 : l-1] {
		parsedLine := strings.Fields(line)
		readKb, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		writeKb, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		e.gzNICUsageRead.With(prometheus.Labels{"device": device}).Set(readKb)
		e.gzNICUsageWrite.With(prometheus.Labels{"device": device}).Set(writeKb)
	}
	return nil
}
