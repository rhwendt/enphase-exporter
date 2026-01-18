package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var productionLog = logrus.WithField("collector", "production")

// ProductionCollector collects production metrics from the Enphase gateway.
type ProductionCollector struct {
	client EnphaseClient

	// Gauges
	productionWatts *prometheus.Desc
	rmsVoltage      *prometheus.Desc
	rmsCurrent      *prometheus.Desc
	powerFactor     *prometheus.Desc

	// Counters
	productionWhTotal *prometheus.Desc
}

// NewProductionCollector creates a new ProductionCollector.
func NewProductionCollector(client EnphaseClient) *ProductionCollector {
	return &ProductionCollector{
		client: client,
		productionWatts: prometheus.NewDesc(
			"enphase_production_watts",
			"Current solar production in watts",
			[]string{"device_type"},
			nil,
		),
		rmsVoltage: prometheus.NewDesc(
			"enphase_production_voltage_volts",
			"RMS voltage",
			[]string{"device_type"},
			nil,
		),
		rmsCurrent: prometheus.NewDesc(
			"enphase_production_current_amps",
			"RMS current in amps",
			[]string{"device_type"},
			nil,
		),
		powerFactor: prometheus.NewDesc(
			"enphase_production_power_factor",
			"Power factor",
			[]string{"device_type"},
			nil,
		),
		productionWhTotal: prometheus.NewDesc(
			"enphase_production_wh_total",
			"Total lifetime production in watt-hours",
			[]string{"device_type"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *ProductionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.productionWatts
	ch <- c.rmsVoltage
	ch <- c.rmsCurrent
	ch <- c.powerFactor
	ch <- c.productionWhTotal
}

// Collect implements prometheus.Collector.
func (c *ProductionCollector) Collect(ch chan<- prometheus.Metric) {
	production, err := c.client.GetProduction()
	if err != nil {
		productionLog.WithError(err).Error("Failed to get production data")
		return
	}

	for _, device := range production.Production {
		ch <- prometheus.MustNewConstMetric(
			c.productionWatts,
			prometheus.GaugeValue,
			device.WNow,
			device.Type,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rmsVoltage,
			prometheus.GaugeValue,
			device.RmsVoltage,
			device.Type,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rmsCurrent,
			prometheus.GaugeValue,
			device.RmsCurrent,
			device.Type,
		)
		ch <- prometheus.MustNewConstMetric(
			c.powerFactor,
			prometheus.GaugeValue,
			device.PwrFactor,
			device.Type,
		)
		ch <- prometheus.MustNewConstMetric(
			c.productionWhTotal,
			prometheus.CounterValue,
			device.WhLifetime,
			device.Type,
		)
	}
}
