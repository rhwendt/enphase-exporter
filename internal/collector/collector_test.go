package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/rhwendt/enphase-exporter/internal/client"
)

// mockClient implements EnphaseClient for testing
type mockClient struct {
	production    *client.ProductionResponse
	inverters     *client.InvertersResponse
	meterReadings *client.MeterReadingsResponse
	err           error
}

func (m *mockClient) GetProduction() (*client.ProductionResponse, error) {
	return m.production, m.err
}

func (m *mockClient) GetInverters() (*client.InvertersResponse, error) {
	return m.inverters, m.err
}

func (m *mockClient) GetMeterReadings() (*client.MeterReadingsResponse, error) {
	return m.meterReadings, m.err
}

func TestProductionCollector(t *testing.T) {
	mock := &mockClient{
		production: &client.ProductionResponse{
			Production: []client.ProductionDevice{
				{
					Type:            "inverters",
					WNow:            2500.5,
					WhLifetime:      1500000,
					WhToday:         5000,
					WhLastSevenDays: 35000,
					RmsVoltage:      240.0,
					RmsCurrent:      10.4,
					PwrFactor:       0.99,
				},
			},
			Consumption: []client.ProductionDevice{
				{
					Type:            "eim",
					MeasurementType: "total-consumption",
					WNow:            1500.0,
					WhLifetime:      3000000,
					WhToday:         10000,
					WhLastSevenDays: 70000,
				},
				{
					Type:            "eim",
					MeasurementType: "net-consumption",
					WNow:            -1000.0,
					WhLifetime:      1500000,
					WhToday:         0,
					WhLastSevenDays: 0,
				},
			},
		},
	}

	collector := NewProductionCollector(mock)

	// Register and collect
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)

	// Test production watts
	expected := `
		# HELP enphase_production_watts Current solar production in watts
		# TYPE enphase_production_watts gauge
		enphase_production_watts{device_type="inverters"} 2500.5
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "enphase_production_watts"); err != nil {
		t.Errorf("production watts mismatch: %v", err)
	}

	// Test consumption watts
	expectedConsumption := `
		# HELP enphase_consumption_watts Current consumption in watts
		# TYPE enphase_consumption_watts gauge
		enphase_consumption_watts{measurement_type="total-consumption"} 1500
		enphase_consumption_watts{measurement_type="net-consumption"} -1000
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expectedConsumption), "enphase_consumption_watts"); err != nil {
		t.Errorf("consumption watts mismatch: %v", err)
	}

	// Test net watts (production - total-consumption = 2500.5 - 1500 = 1000.5)
	expectedNet := `
		# HELP enphase_net_watts Net power (production - consumption). Positive = exporting, negative = importing
		# TYPE enphase_net_watts gauge
		enphase_net_watts 1000.5
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expectedNet), "enphase_net_watts"); err != nil {
		t.Errorf("net watts mismatch: %v", err)
	}
}

func TestInvertersCollector(t *testing.T) {
	mock := &mockClient{
		inverters: &client.InvertersResponse{
			{
				SerialNumber:    "INV001",
				LastReportDate:  1704067200,
				LastReportWatts: 250,
				MaxReportWatts:  300,
			},
			{
				SerialNumber:    "INV002",
				LastReportDate:  1704067200,
				LastReportWatts: 245,
				MaxReportWatts:  300,
			},
		},
	}

	collector := NewInvertersCollector(mock)

	expected := `
		# HELP enphase_inverter_watts Current inverter production in watts
		# TYPE enphase_inverter_watts gauge
		enphase_inverter_watts{serial_number="INV001"} 250
		enphase_inverter_watts{serial_number="INV002"} 245
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "enphase_inverter_watts"); err != nil {
		t.Errorf("inverter watts mismatch: %v", err)
	}

	expectedMax := `
		# HELP enphase_inverter_max_watts Maximum reported inverter production in watts
		# TYPE enphase_inverter_max_watts gauge
		enphase_inverter_max_watts{serial_number="INV001"} 300
		enphase_inverter_max_watts{serial_number="INV002"} 300
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expectedMax), "enphase_inverter_max_watts"); err != nil {
		t.Errorf("inverter max watts mismatch: %v", err)
	}
}

func TestMetersCollector(t *testing.T) {
	mock := &mockClient{
		meterReadings: &client.MeterReadingsResponse{
			{
				Eid:         12345,
				Voltage:     240.5,
				Current:     10.3,
				ActivePower: 2450.0,
				Freq:        60.0,
				PwrFactor:   0.99,
				Channels: []client.MeterChannel{
					{Voltage: 120.2, Current: 5.1, ActivePower: 612.0, PwrFactor: 0.98},
					{Voltage: 120.3, Current: 5.2, ActivePower: 625.0, PwrFactor: 0.99},
				},
			},
		},
	}

	collector := NewMetersCollector(mock)

	// Test frequency (only total, not per-phase)
	expected := `
		# HELP enphase_frequency_hz Grid frequency in Hz
		# TYPE enphase_frequency_hz gauge
		enphase_frequency_hz{meter_id="12345"} 60
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "enphase_frequency_hz"); err != nil {
		t.Errorf("frequency mismatch: %v", err)
	}

	// Test voltage (total and per-phase)
	expectedVoltage := `
		# HELP enphase_voltage_volts Grid voltage in volts
		# TYPE enphase_voltage_volts gauge
		enphase_voltage_volts{meter_id="12345",phase="total"} 240.5
		enphase_voltage_volts{meter_id="12345",phase="L1"} 120.2
		enphase_voltage_volts{meter_id="12345",phase="L2"} 120.3
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expectedVoltage), "enphase_voltage_volts"); err != nil {
		t.Errorf("voltage mismatch: %v", err)
	}
}

func TestCollector_NilData(t *testing.T) {
	// Test that collectors handle nil data gracefully
	mock := &mockClient{
		production:    nil,
		inverters:     nil,
		meterReadings: nil,
	}

	prodCollector := NewProductionCollector(mock)
	invCollector := NewInvertersCollector(mock)
	meterCollector := NewMetersCollector(mock)

	// These should not panic
	ch := make(chan prometheus.Metric, 100)
	prodCollector.Collect(ch)
	invCollector.Collect(ch)
	meterCollector.Collect(ch)
}
