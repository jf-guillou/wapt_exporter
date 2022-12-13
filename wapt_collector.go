package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type WaptCollector struct {
	up     *prometheus.Desc
	hosts  *prometheus.Desc
	online *prometheus.Desc
}

func NewWaptCollector() *WaptCollector {
	return &WaptCollector{
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Is exporter running", nil, nil,
		),
		hosts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "hosts"),
			"Total registered hosts", nil, nil,
		),
		online: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "online"),
			"Online hosts", nil, nil,
		),
	}
}

func (c *WaptCollector) Collect(ch chan<- prometheus.Metric) {
	up := isWaptUp(*waptApi)
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)

	if up == 0 {
		return
	}

	hosts := waptHosts(*waptApi, *waptUser, *waptPassword)
	if hosts == nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(c.hosts, prometheus.GaugeValue, float64(len(hosts.Result)))

	onlineCount := 0
	for _, h := range hosts.Result {
		if h.Reachable == "OK" {
			onlineCount += 1
		}
	}
	ch <- prometheus.MustNewConstMetric(c.online, prometheus.GaugeValue, float64(onlineCount))
}

func (c *WaptCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.hosts
	ch <- c.online
}
