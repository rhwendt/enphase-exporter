package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var productionLog = logrus.WithField("collector", "production")

// ProductionCollector collects production and consumption metrics from the Enphase gateway.
type ProductionCollector struct {
	client EnphaseClient

	// Production gauges
	productionWatts *prometheus.Desc
	rmsVoltage      *prometheus.Desc
	rmsCurrent      *prometheus.Desc
	powerFactor     *prometheus.Desc

	// Production counters
	productionWhTotal *prometheus.Desc

	// Consumption gauges
	consumptionWatts *prometheus.Desc

	// Consumption counters
	consumptionWhTotal *prometheus.Desc

	// Net (production - consumption)
	netWatts *prometheus.Desc
}

// NewProductionCollector creates a new ProductionCollector.
func NewProductionCollector(client EnphaseClient) *ProductionCollector {
	return &ProductionCollector{
		client: client,
		// Production metrics
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
		// Consumption metrics (labeled by measurement_type: "total-consumption" or "net-consumption")
		consumptionWatts: prometheus.NewDesc(
			"enphase_consumption_watts",
			"Current consumption in watts",
			[]string{"measurement_type"},
			nil,
		),
		consumptionWhTotal: prometheus.NewDesc(
			"enphase_consumption_wh_total",
			"Total lifetime consumption in watt-hours",
			[]string{"measurement_type"},
			nil,
		),
		// Net metrics
		netWatts: prometheus.NewDesc(
			"enphase_net_watts",
			"Net power (production - consumption). Positive = exporting, negative = importing",
			nil,
			nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *ProductionCollector) Describe(ch chan<- *prometheus.Desc) {
	// Production
	ch <- c.productionWatts
	ch <- c.rmsVoltage
	ch <- c.rmsCurrent
	ch <- c.powerFactor
	ch <- c.productionWhTotal
	// Consumption
	ch <- c.consumptionWatts
	ch <- c.consumptionWhTotal
	// Net
	ch <- c.netWatts
}

// Collect implements prometheus.Collector.
func (c *ProductionCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch production report
	start := time.Now()
	prodReport, err := c.client.GetProductionReport()
	duration := time.Since(start)
	APICallDuration.WithLabelValues("production_report").Observe(duration.Seconds())
	productionLog.WithField("duration_ms", duration.Milliseconds()).Debug("GetProductionReport completed")
	if err != nil {
		productionLog.WithError(err).Error("Failed to get production report")
		return
	}

	// Fetch consumption report
	start = time.Now()
	consReport, err := c.client.GetConsumptionReport()
	duration = time.Since(start)
	APICallDuration.WithLabelValues("consumption_report").Observe(duration.Seconds())
	productionLog.WithField("duration_ms", duration.Milliseconds()).Debug("GetConsumptionReport completed")
	if err != nil {
		productionLog.WithError(err).Error("Failed to get consumption report")
		return
	}

	if prodReport == nil || consReport == nil {
		return
	}

	// Production metrics from cumulative data
	prod := prodReport.Cumulative
	ch <- prometheus.MustNewConstMetric(
		c.productionWatts,
		prometheus.GaugeValue,
		prod.CurrW,
		"eim",
	)
	ch <- prometheus.MustNewConstMetric(
		c.rmsVoltage,
		prometheus.GaugeValue,
		prod.RmsVoltage,
		"eim",
	)
	ch <- prometheus.MustNewConstMetric(
		c.rmsCurrent,
		prometheus.GaugeValue,
		prod.RmsCurrent,
		"eim",
	)
	ch <- prometheus.MustNewConstMetric(
		c.powerFactor,
		prometheus.GaugeValue,
		prod.PwrFactor,
		"eim",
	)

	// Use lines[] summation for whDlvdCum to fix split-phase doubling
	prodWhTotal := prod.WhDlvdCum
	if len(prodReport.Lines) > 0 {
		prodWhTotal = 0
		for _, line := range prodReport.Lines {
			prodWhTotal += line.WhDlvdCum
		}
	}
	ch <- prometheus.MustNewConstMetric(
		c.productionWhTotal,
		prometheus.CounterValue,
		prodWhTotal,
		"eim",
	)

	// Consumption metrics
	var totalConsumptionW float64
	for _, report := range *consReport {
		label := report.ReportType

		if report.ReportType == "total-consumption" {
			totalConsumptionW = report.Cumulative.CurrW
		}

		ch <- prometheus.MustNewConstMetric(
			c.consumptionWatts,
			prometheus.GaugeValue,
			report.Cumulative.CurrW,
			label,
		)

		// Use lines[] summation for whDlvdCum to fix split-phase doubling
		whTotal := report.Cumulative.WhDlvdCum
		if len(report.Lines) > 0 {
			whTotal = 0
			for _, line := range report.Lines {
				whTotal += line.WhDlvdCum
			}
		}
		ch <- prometheus.MustNewConstMetric(
			c.consumptionWhTotal,
			prometheus.CounterValue,
			whTotal,
			label,
		)
	}

	// Net power (positive = exporting to grid, negative = importing from grid)
	ch <- prometheus.MustNewConstMetric(
		c.netWatts,
		prometheus.GaugeValue,
		prod.CurrW-totalConsumptionW,
	)
}
