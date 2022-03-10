// zpool collector
// this will :
//  - call zpool list
//  - gather ZPOOL metrics
//  - feed the collector

package collector

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// GZZpoolListCollector declares the data type within the prometheus metrics package.
type GZZpoolListCollector struct {
	pools               []string
	gzZpoolListAlloc    *prometheus.GaugeVec
	gzZpoolListCapacity *prometheus.GaugeVec
	gzZpoolListFaulty   *prometheus.GaugeVec
	gzZpoolListFrag     *prometheus.GaugeVec
	gzZpoolListFree     *prometheus.GaugeVec
	gzZpoolListSize     *prometheus.GaugeVec
}

// NewGZZpoolListExporter returns a newly allocated exporter GZZpoolListCollector.
// It exposes the zpool list command result.
func NewGZZpoolListExporter(pools ...string) (*GZZpoolListCollector, error) {
	if len(pools) < 1 {
		return nil, errors.New("No pools provided")
	}
	return &GZZpoolListCollector{
		pools: pools,
		gzZpoolListAlloc: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_alloc_bytes",
			Help: "ZFS zpool allocated size in bytes.",
		}, []string{"zpool"}),
		gzZpoolListCapacity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_cap_percents",
			Help: "ZFS zpool capacity in percents.",
		}, []string{"zpool"}),
		gzZpoolListFaulty: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_faults",
			Help: "ZFS zpool health status.",
		}, []string{"zpool"}),
		gzZpoolListFrag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_frag_percents",
			Help: "ZFS zpool fragmentation in percents.",
		}, []string{"zpool"}),
		gzZpoolListFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_free_bytes",
			Help: "ZFS zpool space available in bytes.",
		}, []string{"zpool"}),
		gzZpoolListSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_size_bytes",
			Help: "ZFS zpool allocated size in bytes.",
		}, []string{"zpool"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZZpoolListCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzZpoolListAlloc.Describe(ch)
	e.gzZpoolListCapacity.Describe(ch)
	e.gzZpoolListFaulty.Describe(ch)
	e.gzZpoolListFrag.Describe(ch)
	e.gzZpoolListFree.Describe(ch)
	e.gzZpoolListSize.Describe(ch)
}

// Collect fetches the stats.
func (e *GZZpoolListCollector) Collect(ch chan<- prometheus.Metric) {
	e.zpoolList()
	e.gzZpoolListAlloc.Collect(ch)
	e.gzZpoolListCapacity.Collect(ch)
	e.gzZpoolListFaulty.Collect(ch)
	e.gzZpoolListFrag.Collect(ch)
	e.gzZpoolListFree.Collect(ch)
	e.gzZpoolListSize.Collect(ch)
}

func (e *GZZpoolListCollector) zpoolList() {
	out, eerr := exec.Command("zpool", append([]string{"list", "-p"}, e.pools...)...).Output()
	if eerr != nil {
		log.Errorf("error on executing zpool: %v", eerr)
	}
	perr := e.parseZpoolListOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing zpool: %v", perr)
	}
}

func (e *GZZpoolListCollector) parseZpoolListOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		// handle different version of zpool output (CKPOINT)
		// lazy version : just shift the variable assignation when needed
		// only two cases are handled currently :
		//	fieldNumber = 10 -> zpool output WITHOUT CKPOINT feature
		//	fieldNumber = 11 -> zpool output WITH CKPOINT feature
		fieldNumber := len(parsedLine)
		n := 0
		if fieldNumber == 11 {
			n = 1
		}
		pool := parsedLine[0]
		sizeBytes, err := strconv.ParseFloat(parsedLine[1], 64)
		if err != nil {
			return err
		}
		allocBytes, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		freeBytes, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		fragPercent := strings.TrimSuffix(parsedLine[5+n], "%")
		fragPercentTrim, err := strconv.ParseFloat(fragPercent, 64)
		if err != nil {
			return err
		}
		capPercent := strings.TrimSuffix(parsedLine[6+n], "%")
		capPercentTrim, err := strconv.ParseFloat(capPercent, 64)
		if err != nil {
			return err
		}
		health := parsedLine[8+n]
		if (strings.Contains(health, "ONLINE")) == true {
			e.gzZpoolListFaulty.With(prometheus.Labels{"zpool": pool}).Set(0)
		} else {
			e.gzZpoolListFaulty.With(prometheus.Labels{"zpool": pool}).Set(1)
		}

		e.gzZpoolListAlloc.With(prometheus.Labels{"zpool": pool}).Set(allocBytes)
		e.gzZpoolListCapacity.With(prometheus.Labels{"zpool": pool}).Set(capPercentTrim)
		e.gzZpoolListFrag.With(prometheus.Labels{"zpool": pool}).Set(fragPercentTrim)
		e.gzZpoolListFree.With(prometheus.Labels{"zpool": pool}).Set(freeBytes)
		e.gzZpoolListSize.With(prometheus.Labels{"zpool": pool}).Set(sizeBytes)
	}
	return nil
}
