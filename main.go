// Virtua SmartOS Prometheus exporter
//
// Worflow :
//  - detect if launched in GZ or into a Zone
//  - retrieve useful metrics
//  - expose them to http://xxx:9100/metrics (same as node_exporter)

package main

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/thetooth/smartos_exporter/collector"

	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Global variables
	listenAddress = kingpin.Flag("server.listen-address", "Address on which to expose metrics and web interface.").Default(":9100").String()
	pools         = kingpin.Flag("gz.pools", "List of zfs pools to monitor. e.g. zones,tank,bread,milk").Required().Strings()
	nics          = kingpin.Flag("gz.nics", "List of network interfaces to monitor. e.g. loop0,ixgbe0,ixgbe1").Required().Strings()
)

func init() {
	// change PATH env variable inside a LX zone
	if runtime.GOOS == "linux" {
		os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/native/usr/bin")
	}
}

// Global Helpers

// try to determine if its executed inside the GZ or not.
// return 1 if in GZ
//        0 if in zone
func isGlobalZone() int {
	out, eerr := exec.Command("zonename").Output()
	if eerr != nil {
		log.Fatal(eerr)
	}
	if (strings.Contains(string(out), "global")) == false {
		return 0
	}
	return 1
}

// program starter
func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("smartos_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting smartos_exporter", version.Info())
	// check if it is a GZ or a zone
	gz := isGlobalZone()

	// common metrics
	loadAvg, _ := collector.NewLoadAverageExporter()
	prometheus.MustRegister(loadAvg)

	if gz == 0 {
		// Zone metrics
		zoneDf, _ := collector.NewZoneDfExporter()
		prometheus.MustRegister(zoneDf)

		zoneKstat, _ := collector.NewZoneKstatExporter()
		prometheus.MustRegister(zoneKstat)
	}

	if gz == 1 {
		// Global Zone metrics
		gzFreeMem, _ := collector.NewGZFreeMemExporter()
		prometheus.MustRegister(gzFreeMem)

		gzNICUsage, _ := collector.NewGZNICUsageExporter(*nics...)
		prometheus.MustRegister(gzNICUsage)

		cpuUsage, _ := collector.NewGZCPUUsageExporter()
		prometheus.MustRegister(cpuUsage)

		gzDiskErrors, _ := collector.NewGZDiskErrorsExporter()
		prometheus.MustRegister(gzDiskErrors)

		gzZpoolList, _ := collector.NewGZZpoolListExporter(*pools...)
		prometheus.MustRegister(gzZpoolList)

		gzSMARTList, _ := collector.NewGZDiskSMARTExporter(*pools...)
		prometheus.MustRegister(gzSMARTList)
	}

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	http.Handle("/metrics", promhttp.Handler())
	log.Infoln("Listening on", *listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
