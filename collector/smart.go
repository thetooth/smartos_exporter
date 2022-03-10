// iostat collector
// this will :
//  - call iostat
//  - gather hard disk metrics
//  - feed the collector

package collector

import (
	"os/exec"
	"strconv"
	"strings"

	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// GZDiskSMARTCollector declares the data type within the prometheus metrics package.
type GZDiskSMARTCollector struct {
	DeviceList        map[string][]string
	gzDiskTemperature *prometheus.GaugeVec
}

// NewGZDiskErrorsExporter returns a newly allocated exporter GZDiskSMARTCollector.
// It exposes the disk temperature as reported by smartctl
func NewGZDiskSMARTExporter(pools ...string) (*GZDiskSMARTCollector, error) {
	devices := map[string][]string{}
	for _, pool := range pools {
		devices[pool] = disksInPool(pool)
	}
	return &GZDiskSMARTCollector{
		gzDiskTemperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_disk_temp",
			Help: "Disk temperature in degrees C.",
		}, []string{"pool", "device"}),
		DeviceList: devices,
	}, nil
}

// Describe describes all the metrics.
func (e *GZDiskSMARTCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzDiskTemperature.Describe(ch)
}

// Collect fetches the stats.
func (e *GZDiskSMARTCollector) Collect(ch chan<- prometheus.Metric) {
	e.smartctl()
	e.gzDiskTemperature.Collect(ch)
}

// smartctl  -d sat,12 /dev/rdsk/c2t2d0 -A
func (e *GZDiskSMARTCollector) smartctl() {
	for pool, devices := range e.DeviceList {
		for _, device := range devices {
			out, eerr := exec.Command("/opt/tools/sbin/smartctl", "-d", "sat,12", "/dev/rdsk/"+device, "-A").Output()
			if eerr != nil {
				log.Errorf("error on executing smartctl: %v", eerr)
			}
			perr := e.parseSmartCtlOutput(pool, device, string(out))
			if perr != nil {
				log.Errorf("error on parsing smartctl: %v", perr)
			}
		}
	}
}

func (e *GZDiskSMARTCollector) parseSmartCtlOutput(pool, deviceName, out string) error {
	outlines := strings.Split(out, "\n")
	for _, line := range outlines {
		if strings.Contains(line, "Temperature_Celsius") || strings.Contains(line, "Temperature_Case") {
			parsedLine := strings.Fields(line)
			celsius, err := strconv.ParseFloat(parsedLine[3], 64)
			if err != nil {
				return err
			}

			e.gzDiskTemperature.With(prometheus.Labels{"pool": pool, "device": deviceName}).Set(celsius)
			return nil
		}
	}
	return nil
}

func disksInPool(pool string) (disks []string) {
	out, eerr := exec.Command("/usr/sbin/zpool", "iostat", "-v", pool).Output()
	if eerr != nil {
		log.Errorf("error on executing smartctl: %v", eerr)
	}
	outlines := strings.Split(string(out), "\n")
	maxLevel := 0
	for index, line := range outlines[3 : len(outlines)-1] {
		if index == 0 && strings.Count(line[:4], " ") != 0 {
			log.Error("Pool name is less than 4 characters, indentation magic be broken...")
			return
		}
		if indent := strings.Count(line[:4], " "); indent > maxLevel {
			maxLevel = indent
		}
	}
	prefix := strings.Repeat(" ", maxLevel)
	for _, line := range outlines[3 : len(outlines)-1] {
		if strings.HasPrefix(line, prefix) {
			fields := strings.Fields(line)
			disks = append(disks, fields[0])
		}
	}

	return
}
