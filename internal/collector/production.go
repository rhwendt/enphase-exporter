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
	productionWhTotal     *prometheus.Desc
	productionWhToday     *prometheus.Desc
	productionWhLastWeek  *prometheus.Desc

	// Consumption gauges
	consumptionWatts *prometheus.Desc

	// Consumption counters
	consumptionWhTotal    *prometheus.Desc
	consumptionWhToday    *prometheus.Desc
	consumptionWhLastWeek *prometheus.Desc

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
		productionWhToday: prometheus.NewDesc(
			"enphase_production_wh_today",
			"Production today in watt-hours (WARNING: gateway value may not reset at midnight - use increase(enphase_production_wh_total[24h]) instead)",
			[]string{"device_type"},
			nil,
		),
		productionWhLastWeek: prometheus.NewDesc(
			"enphase_production_wh_last_seven_days",
			"Production in last 7 days in watt-hours",
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
		consumptionWhToday: prometheus.NewDesc(
			"enphase_consumption_wh_today",
			"Consumption today in watt-hours (WARNING: gateway value may not reset at midnight - use increase(enphase_consumption_wh_total[24h]) instead)",
			[]string{"measurement_type"},
			nil,
		),
		consumptionWhLastWeek: prometheus.NewDesc(
			"enphase_consumption_wh_last_seven_days",
			"Consumption in last 7 days in watt-hours",
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
	ch <- c.productionWhToday
	ch <- c.productionWhLastWeek
	// Consumption
	ch <- c.consumptionWatts
	ch <- c.consumptionWhTotal
	ch <- c.consumptionWhToday
	ch <- c.consumptionWhLastWeek
	// Net
	ch <- c.netWatts
}

// Collect implements prometheus.Collector.
func (c *ProductionCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	production, err := c.client.GetProduction()
	duration := time.Since(start)
	APICallDuration.WithLabelValues("production").Observe(duration.Seconds())
	productionLog.WithField("duration_ms", duration.Milliseconds()).Debug("GetProduction completed")
	if err != nil {
		productionLog.WithError(err).Error("Failed to get production data")
		return
	}

	if production == nil {
		return
	}

	var totalProduction, totalConsumption float64

	// Production metrics
	for _, device := range production.Production {
		totalProduction += device.WNow

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
		ch <- prometheus.MustNewConstMetric(
			c.productionWhToday,
			prometheus.GaugeValue,
			device.WhToday,
			device.Type,
		)
		ch <- prometheus.MustNewConstMetric(
			c.productionWhLastWeek,
			prometheus.GaugeValue,
			device.WhLastSevenDays,
			device.Type,
		)
	}

	// Consumption metrics (use MeasurementType to distinguish total-consumption vs net-consumption)
	//
	// KNOWN BUG - Enphase Gateway Daily Metrics:
	// The Enphase gateway has bugs affecting daily/weekly metrics:
	// 1. Split-phase systems: Top-level whToday/whLastSevenDays/whLifetime values are doubled
	// 2. All systems: The whToday values may NOT reset at midnight (accumulates over days)
	//
	// We use lines[] summation which fixes the doubling issue (#1).
	// However, the midnight reset issue (#2) is a gateway firmware bug we cannot fix.
	//
	// WORKAROUND - Use Prometheus queries on the lifetime counter instead:
	//
	//   Rolling 24-hour:
	//     increase(enphase_production_wh_total{device_type="eim"}[24h])
	//     increase(enphase_consumption_wh_total{measurement_type="total-consumption"}[24h])
	//
	//   Today so far (in Grafana, set time range to "Today so far"):
	//     increase(enphase_production_wh_total{device_type="eim"}[$__range])
	//     increase(enphase_consumption_wh_total{measurement_type="total-consumption"}[$__range])
	//
	// The _wh_total metrics are lifetime counters that always increase, making them
	// reliable for calculating daily values via Prometheus increase() function.
	for _, device := range production.Consumption {
		// Use MeasurementType if available, otherwise fall back to Type
		label := device.MeasurementType
		if label == "" {
			label = device.Type
		}

		// Only use total-consumption for the net power calculation
		if device.MeasurementType == "total-consumption" {
			totalConsumption = device.WNow
		}

		// Use lines[] summation to avoid split-phase doubling (see KNOWN LIMITATION above)
		whToday := device.WhToday
		whLastSevenDays := device.WhLastSevenDays
		whLifetime := device.WhLifetime
		if len(device.Lines) > 0 {
			whToday = 0
			whLastSevenDays = 0
			whLifetime = 0
			for _, line := range device.Lines {
				whToday += line.WhToday
				whLastSevenDays += line.WhLastSevenDays
				whLifetime += line.WhLifetime
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.consumptionWatts,
			prometheus.GaugeValue,
			device.WNow,
			label,
		)
		ch <- prometheus.MustNewConstMetric(
			c.consumptionWhTotal,
			prometheus.CounterValue,
			whLifetime,
			label,
		)
		ch <- prometheus.MustNewConstMetric(
			c.consumptionWhToday,
			prometheus.GaugeValue,
			whToday,
			label,
		)
		ch <- prometheus.MustNewConstMetric(
			c.consumptionWhLastWeek,
			prometheus.GaugeValue,
			whLastSevenDays,
			label,
		)
	}

	// Net power (positive = exporting to grid, negative = importing from grid)
	ch <- prometheus.MustNewConstMetric(
		c.netWatts,
		prometheus.GaugeValue,
		totalProduction-totalConsumption,
	)
}
