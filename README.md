# smartos_exporter
[![Go Report Card](https://goreportcard.com/badge/github.com/thetooth/smartos_exporter)](https://goreportcard.com/report/github.com/thetooth/smartos_exporter)

Golang program for gathering SmartOS statistics and providing them to Prometheus

It relies on SmartOS tools and commands so it should be very easy to implement a
new probe.

This fork adds some more cli options so it can be reconfigured at runtime.

```
usage: main --gz.pools=GZ.POOLS --gz.nics=GZ.NICS [<flags>]

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
      --server.listen-address=":9100"  
                               Address on which to expose metrics and web interface.
      --gz.pools=GZ.POOLS ...  List of zfs pools to monitor. e.g. zones,tank,bread,milk
      --gz.nics=GZ.NICS ...    List of network interfaces to monitor. e.g. loop0,ixgbe0,ixgbe1
      --log.level="info"       Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"  
                               Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version                Show application version.
```

```
# cp ./smartos_exporter /opt/custom/smf/bin/smartos_exporter
# svccfg import smartos_exporter.xml
# svccfg -s smartos_exporter
svc:/smartos_exporter> setprop gz/nics=ixgbe0,loop0
svc:/smartos_exporter> setprop gz/pools=zones
svc:/smartos_exporter> end
# svccfg -s smartos_exporter:default refresh
# svcadm restart smartos_exporter
```
