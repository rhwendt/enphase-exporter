package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var metersLog = logrus.WithField("collector", "meters")

const meterTypeRefreshInterval = 15 * time.Minute

// MetersCollector collects meter readings from the Enphase gateway.
type MetersCollector struct {
	client EnphaseClient

	// Cached meter metadata
	meterTypes   map[int64]string
	meterTypesMu sync.RWMutex
	lastRefresh  time.Time

	voltage         *prometheus.Desc
	current         *prometheus.Desc
	activePower     *prometheus.Desc
	powerFactor     *prometheus.Desc
	frequency       *prometheus.Desc
	energyDelivered *prometheus.Desc
	energyReceived  *prometheus.Desc
}

// NewMetersCollector creates a new MetersCollector.
func NewMetersCollector(client EnphaseClient) *MetersCollector {
	c := &MetersCollector{
		client:     client,
		meterTypes: make(map[int64]string),
		voltage: prometheus.NewDesc(
			"enphase_voltage_volts",
			"Grid voltage in volts",
			[]string{"measurement_type", "meter_id", "phase"},
			nil,
		),
		current: prometheus.NewDesc(
			"enphase_current_amps",
			"Current in amps",
			[]string{"measurement_type", "meter_id", "phase"},
			nil,
		),
		activePower: prometheus.NewDesc(
			"enphase_active_power_watts",
			"Active power in watts",
			[]string{"measurement_type", "meter_id", "phase"},
			nil,
		),
		powerFactor: prometheus.NewDesc(
			"enphase_power_factor",
			"Power factor",
			[]string{"measurement_type", "meter_id", "phase"},
			nil,
		),
		frequency: prometheus.NewDesc(
			"enphase_frequency_hz",
			"Grid frequency in Hz",
			[]string{"measurement_type", "meter_id"},
			nil,
		),
		energyDelivered: prometheus.NewDesc(
			"enphase_meter_energy_delivered_wh",
			"Cumulative energy delivered in watt-hours (meaning depends on measurement_type)",
			[]string{"measurement_type", "meter_id", "phase"},
			nil,
		),
		energyReceived: prometheus.NewDesc(
			"enphase_meter_energy_received_wh",
			"Cumulative energy received in watt-hours (meaning depends on measurement_type)",
			[]string{"measurement_type", "meter_id", "phase"},
			nil,
		),
	}
	c.refreshMeterTypes()
	return c
}

func (c *MetersCollector) refreshMeterTypes() {
	meters, err := c.client.GetMeters()
	if err != nil {
		metersLog.WithError(err).Warn("Failed to fetch meter metadata, will retry on next scrape")
		return
	}
	if meters == nil {
		return
	}
	c.meterTypesMu.Lock()
	defer c.meterTypesMu.Unlock()
	for _, m := range *meters {
		c.meterTypes[m.Eid] = m.MeasurementType
		metersLog.WithFields(logrus.Fields{
			"eid":   m.Eid,
			"type":  m.MeasurementType,
			"state": m.State,
		}).Info("Discovered meter")
	}
	c.lastRefresh = time.Now()
}

func (c *MetersCollector) getMeasurementType(eid int64) string {
	c.meterTypesMu.RLock()
	defer c.meterTypesMu.RUnlock()
	if mt, ok := c.meterTypes[eid]; ok {
		return mt
	}
	return "unknown"
}

// Describe implements prometheus.Collector.
func (c *MetersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.voltage
	ch <- c.current
	ch <- c.activePower
	ch <- c.powerFactor
	ch <- c.frequency
	ch <- c.energyDelivered
	ch <- c.energyReceived
}

// Collect implements prometheus.Collector.
func (c *MetersCollector) Collect(ch chan<- prometheus.Metric) {
	// Refresh meter metadata if stale or empty
	c.meterTypesMu.RLock()
	needsRefresh := time.Since(c.lastRefresh) > meterTypeRefreshInterval || len(c.meterTypes) == 0
	c.meterTypesMu.RUnlock()
	if needsRefresh {
		c.refreshMeterTypes()
	}

	start := time.Now()
	readings, err := c.client.GetMeterReadings()
	duration := time.Since(start)
	APICallDuration.WithLabelValues("meters").Observe(duration.Seconds())
	metersLog.WithField("duration_ms", duration.Milliseconds()).Debug("GetMeterReadings completed")
	if err != nil {
		metersLog.WithError(err).Error("Failed to get meter readings")
		return
	}

	if readings == nil {
		return
	}

	for _, reading := range *readings {
		meterID := fmt.Sprintf("%d", reading.Eid)
		measurementType := c.getMeasurementType(reading.Eid)

		// Total meter values
		ch <- prometheus.MustNewConstMetric(
			c.voltage,
			prometheus.GaugeValue,
			reading.Voltage,
			measurementType, meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.current,
			prometheus.GaugeValue,
			reading.Current,
			measurementType, meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.activePower,
			prometheus.GaugeValue,
			reading.ActivePower,
			measurementType, meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.powerFactor,
			prometheus.GaugeValue,
			reading.PwrFactor,
			measurementType, meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.frequency,
			prometheus.GaugeValue,
			reading.Freq,
			measurementType, meterID,
		)
		ch <- prometheus.MustNewConstMetric(
			c.energyDelivered,
			prometheus.CounterValue,
			reading.ActEnergyDlvd,
			measurementType, meterID, "total",
		)
		ch <- prometheus.MustNewConstMetric(
			c.energyReceived,
			prometheus.CounterValue,
			reading.ActEnergyRcvd,
			measurementType, meterID, "total",
		)

		// Per-phase values
		for i, channel := range reading.Channels {
			phase := fmt.Sprintf("L%d", i+1)
			ch <- prometheus.MustNewConstMetric(
				c.voltage,
				prometheus.GaugeValue,
				channel.Voltage,
				measurementType, meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.current,
				prometheus.GaugeValue,
				channel.Current,
				measurementType, meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.activePower,
				prometheus.GaugeValue,
				channel.ActivePower,
				measurementType, meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.powerFactor,
				prometheus.GaugeValue,
				channel.PwrFactor,
				measurementType, meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.energyDelivered,
				prometheus.CounterValue,
				channel.ActEnergyDlvd,
				measurementType, meterID, phase,
			)
			ch <- prometheus.MustNewConstMetric(
				c.energyReceived,
				prometheus.CounterValue,
				channel.ActEnergyRcvd,
				measurementType, meterID, phase,
			)
		}
	}
}
