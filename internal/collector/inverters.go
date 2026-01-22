package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var invertersLog = logrus.WithField("collector", "inverters")

// InvertersCollector collects per-inverter metrics from the Enphase gateway.
type InvertersCollector struct {
	client EnphaseClient

	inverterWatts    *prometheus.Desc
	inverterMaxWatts *prometheus.Desc
	inverterLastReport *prometheus.Desc
}

// NewInvertersCollector creates a new InvertersCollector.
func NewInvertersCollector(client EnphaseClient) *InvertersCollector {
	return &InvertersCollector{
		client: client,
		inverterWatts: prometheus.NewDesc(
			"enphase_inverter_watts",
			"Current inverter production in watts",
			[]string{"serial_number"},
			nil,
		),
		inverterMaxWatts: prometheus.NewDesc(
			"enphase_inverter_max_watts",
			"Maximum reported inverter production in watts",
			[]string{"serial_number"},
			nil,
		),
		inverterLastReport: prometheus.NewDesc(
			"enphase_inverter_last_report_timestamp",
			"Unix timestamp of last inverter report",
			[]string{"serial_number"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *InvertersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.inverterWatts
	ch <- c.inverterMaxWatts
	ch <- c.inverterLastReport
}

// Collect implements prometheus.Collector.
func (c *InvertersCollector) Collect(ch chan<- prometheus.Metric) {
	inverters, err := c.client.GetInverters()
	if err != nil {
		invertersLog.WithError(err).Error("Failed to get inverter data")
		return
	}

	if inverters == nil {
		return
	}

	for _, inv := range *inverters {
		ch <- prometheus.MustNewConstMetric(
			c.inverterWatts,
			prometheus.GaugeValue,
			float64(inv.LastReportWatts),
			inv.SerialNumber,
		)
		ch <- prometheus.MustNewConstMetric(
			c.inverterMaxWatts,
			prometheus.GaugeValue,
			float64(inv.MaxReportWatts),
			inv.SerialNumber,
		)
		ch <- prometheus.MustNewConstMetric(
			c.inverterLastReport,
			prometheus.GaugeValue,
			float64(inv.LastReportDate),
			inv.SerialNumber,
		)
	}
}
