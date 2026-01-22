package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var metersLog = logrus.WithField("collector", "meters")

// MetersCollector collects meter readings from the Enphase gateway.
type MetersCollector struct {
	client EnphaseClient

	voltage        *prometheus.Desc
	current        *prometheus.Desc
	activePower    *prometheus.Desc
	powerFactor    *prometheus.Desc
	frequency      *prometheus.Desc
	energyExported *prometheus.Desc
	energyImported *prometheus.Desc
}

// NewMetersCollector creates a new MetersCollector.
func NewMetersCollector(client EnphaseClient) *MetersCollector {
	return &MetersCollector{
		client: client,
		voltage: prometheus.NewDesc(
			"enphase_voltage_volts",
			"Grid voltage in volts",
			[]string{"meter_id", "phase"},
			nil,
		),
		current: prometheus.NewDesc(
			"enphase_current_amps",
			"Current in amps",
			[]string{"meter_id", "phase"},
			nil,
		),
		activePower: prometheus.NewDesc(
			"enphase_active_power_watts",
			"Active power in watts",
			[]string{"meter_id", "phase"},
			nil,
		),
		powerFactor: prometheus.NewDesc(
			"enphase_power_factor",
			"Power factor",
			[]string{"meter_id", "phase"},
			nil,
		),
		frequency: prometheus.NewDesc(
			"enphase_frequency_hz",
			"Grid frequency in Hz",
			[]string{"meter_id"},
			nil,
		),
		energyExported: prometheus.NewDesc(
			"enphase_energy_exported_wh",
			"Cumulative energy exported to grid in watt-hours",
			[]string{"meter_id", "phase"},
			nil,
		),
		energyImported: prometheus.NewDesc(
			"enphase_energy_imported_wh",
			"Cumulative energy imported from grid in watt-hours",
			[]string{"meter_id", "phase"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *MetersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.voltage
	ch <- c.current
	ch <- c.activePower
	ch <- c.powerFactor
	ch <- c.frequency
	ch <- c.energyExported
	ch <- c.energyImported
}

// Collect implements prometheus.Collector.
func (c *MetersCollector) Collect(ch chan<- prometheus.Metric) {
	readings, err := c.client.GetMeterReadings()
	if err != nil {
		metersLog.WithError(err).Error("Failed to get meter readings")
		return
	}

	if readings == nil {
		return
	}

	for _, reading := range *readings {
		meterID := fmt.Sprintf("%d", reading.Eid)

		// Total meter values
		ch <- prometheus.MustNewConstMetric(
			c.voltage,
			prometheus.GaugeValue,
			reading.Voltage,
			meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.current,
			prometheus.GaugeValue,
			reading.Current,
			meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.activePower,
			prometheus.GaugeValue,
			reading.ActivePower,
			meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.powerFactor,
			prometheus.GaugeValue,
			reading.PwrFactor,
			meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.frequency,
			prometheus.GaugeValue,
			reading.Freq,
			meterID,
		)
		ch <- prometheus.MustNewConstMetric(
			c.energyExported,
			prometheus.CounterValue,
			reading.ActEnergyDlvd,
			meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.energyImported,
			prometheus.CounterValue,
			reading.ActEnergyRcvd,
			meterID, "total",
		)

		// Per-phase values
		for i, channel := range reading.Channels {
			phase := fmt.Sprintf("L%d", i+1)
			ch <- prometheus.MustNewConstMetric(
				c.voltage,
				prometheus.GaugeValue,
				channel.Voltage,
				meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.current,
				prometheus.GaugeValue,
				channel.Current,
				meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.activePower,
				prometheus.GaugeValue,
				channel.ActivePower,
				meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.powerFactor,
				prometheus.GaugeValue,
				channel.PwrFactor,
				meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.energyExported,
				prometheus.CounterValue,
				channel.ActEnergyDlvd,
				meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.energyImported,
				prometheus.CounterValue,
				channel.ActEnergyRcvd,
				meterID, phase,
			)
		}
	}
}
