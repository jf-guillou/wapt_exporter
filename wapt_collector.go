package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type WaptCollector struct {
	up       *prometheus.Desc
	hosts    *prometheus.Desc
	packages *prometheus.Desc
}

func NewWaptCollector() *WaptCollector {
	return &WaptCollector{
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Is exporter running", nil, nil,
		),
		hosts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "hosts"),
			"Registered hosts", []string{"reachable", "version"}, nil,
		),
		packages: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "packages"),
			"Local packages", nil, nil,
		),
	}
}

func (c *WaptCollector) Collect(ch chan<- prometheus.Metric) {
	up := isWaptUp(*waptApi)
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)

	if up == 0 {
		return
	}

	// Hosts
	hosts := waptHosts(*waptApi, *waptUser, *waptPassword)
	if hosts == nil {
		return
	}

	vmap := make(map[string]map[string]float64)
	for _, h := range hosts.Result {
		if vmap[h.WaptVersion] == nil {
			vmap[h.WaptVersion] = make(map[string]float64)
		}
		vmap[h.WaptVersion][h.Reachable] += 1
	}

	for version, reachableStates := range vmap {
		for reachableState, count := range reachableStates {
			ch <- prometheus.MustNewConstMetric(c.hosts, prometheus.GaugeValue, count, reachableState, version)
		}
	}

	// Packages
	packages := waptPackages(*waptApi, *waptUser, *waptPassword)
	if packages == nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(c.packages, prometheus.GaugeValue, float64(len(packages.Result)))
}

func (c *WaptCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.hosts
	ch <- c.packages
}
