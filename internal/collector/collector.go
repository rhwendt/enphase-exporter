package collector

import "github.com/rhwendt/enphase-exporter/internal/client"

// EnphaseClient defines the interface for the Enphase client used by collectors.
type EnphaseClient interface {
	GetProduction() (*client.ProductionResponse, error)
	GetMeterReadings() (*client.MeterReadingsResponse, error)
	GetInverters() (*client.InvertersResponse, error)
}
