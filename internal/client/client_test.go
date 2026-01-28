package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with JWT",
			config: Config{
				Address: "https://envoy.local",
				Serial:  "123456789",
				JWT:     "valid-jwt-token",
			},
			wantErr: false,
		},
		{
			name: "missing address",
			config: Config{
				Serial: "123456789",
				JWT:    "valid-jwt-token",
			},
			wantErr: true,
		},
		{
			name: "missing serial",
			config: Config{
				Address: "https://envoy.local",
				JWT:     "valid-jwt-token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetProductionReport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/check_jwt":
			w.WriteHeader(http.StatusOK)
		case "/ivp/meters/reports/production":
			resp := ProductionReportResponse{
				CreatedAt:  1706400000,
				ReportType: "production",
				Cumulative: MeterReportData{
					CurrW:      2500.5,
					WhDlvdCum:  1500000,
					RmsVoltage: 240.5,
					RmsCurrent: 10.3,
					PwrFactor:  0.99,
				},
			}
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(Config{
		Address: server.URL,
		Serial:  "123456789",
		JWT:     "test-jwt",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.httpClient = server.Client()

	report, err := client.GetProductionReport()
	if err != nil {
		t.Fatalf("GetProductionReport() error = %v", err)
	}

	if report.Cumulative.CurrW != 2500.5 {
		t.Errorf("Expected production watts 2500.5, got %f", report.Cumulative.CurrW)
	}

	if report.ReportType != "production" {
		t.Errorf("Expected report type 'production', got '%s'", report.ReportType)
	}
}

func TestClient_GetConsumptionReport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/check_jwt":
			w.WriteHeader(http.StatusOK)
		case "/ivp/meters/reports/consumption":
			resp := ConsumptionReportResponse{
				{
					CreatedAt:  1706400000,
					ReportType: "total-consumption",
					Cumulative: MeterReportData{
						CurrW:     1500.0,
						WhDlvdCum: 3000000,
					},
				},
				{
					CreatedAt:  1706400000,
					ReportType: "net-consumption",
					Cumulative: MeterReportData{
						CurrW:     -1000.0,
						WhDlvdCum: 1500000,
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(Config{
		Address: server.URL,
		Serial:  "123456789",
		JWT:     "test-jwt",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.httpClient = server.Client()

	report, err := client.GetConsumptionReport()
	if err != nil {
		t.Fatalf("GetConsumptionReport() error = %v", err)
	}

	if len(*report) != 2 {
		t.Fatalf("Expected 2 consumption reports, got %d", len(*report))
	}

	if (*report)[0].ReportType != "total-consumption" {
		t.Errorf("Expected report type 'total-consumption', got '%s'", (*report)[0].ReportType)
	}

	if (*report)[0].Cumulative.CurrW != 1500.0 {
		t.Errorf("Expected consumption watts 1500.0, got %f", (*report)[0].Cumulative.CurrW)
	}
}

func TestClient_GetInverters(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/check_jwt":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/production/inverters":
			resp := InvertersResponse{
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
			}
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(Config{
		Address: server.URL,
		Serial:  "123456789",
		JWT:     "test-jwt",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.httpClient = server.Client()

	inverters, err := client.GetInverters()
	if err != nil {
		t.Fatalf("GetInverters() error = %v", err)
	}

	if inverters == nil {
		t.Fatal("Expected inverters, got nil")
	}

	if len(*inverters) != 2 {
		t.Errorf("Expected 2 inverters, got %d", len(*inverters))
	}

	if (*inverters)[0].SerialNumber != "INV001" {
		t.Errorf("Expected serial 'INV001', got '%s'", (*inverters)[0].SerialNumber)
	}
}

func TestClient_GetMeterReadings(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/check_jwt":
			w.WriteHeader(http.StatusOK)
		case "/ivp/meters/readings":
			resp := MeterReadingsResponse{
				{
					Eid:         12345,
					Voltage:     240.5,
					Current:     10.3,
					ActivePower: 2450.0,
					Freq:        60.0,
					PwrFactor:   0.99,
					Channels: []MeterChannel{
						{Voltage: 120.2, Current: 5.1, ActivePower: 612.0},
						{Voltage: 120.3, Current: 5.2, ActivePower: 625.0},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(Config{
		Address: server.URL,
		Serial:  "123456789",
		JWT:     "test-jwt",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.httpClient = server.Client()

	readings, err := client.GetMeterReadings()
	if err != nil {
		t.Fatalf("GetMeterReadings() error = %v", err)
	}

	if readings == nil {
		t.Fatal("Expected meter readings, got nil")
	}

	if len(*readings) != 1 {
		t.Errorf("Expected 1 meter, got %d", len(*readings))
	}

	meter := (*readings)[0]
	if meter.Voltage != 240.5 {
		t.Errorf("Expected voltage 240.5, got %f", meter.Voltage)
	}

	if len(meter.Channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(meter.Channels))
	}
}

func TestClient_IsReady(t *testing.T) {
	client := &Client{
		ready: false,
	}

	if client.IsReady() {
		t.Error("Expected IsReady() to return false when not ready")
	}

	// Simulate authenticated state
	client.ready = true
	client.sessionID = "test-session"
	client.sessionExp = time.Now().Add(30 * time.Minute)

	if !client.IsReady() {
		t.Error("Expected IsReady() to return true when authenticated")
	}

	// Test expired session
	client.sessionExp = time.Now().Add(-1 * time.Minute)
	if client.IsReady() {
		t.Error("Expected IsReady() to return false when session expired")
	}
}
