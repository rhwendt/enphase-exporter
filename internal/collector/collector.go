package collector

import "github.com/rhwendt/enphase-exporter/internal/client"

// EnphaseClient defines the interface for the Enphase client used by collectors.
type EnphaseClient interface {
	GetProductionReport() (*client.ProductionReportResponse, error)
	GetConsumptionReport() (*client.ConsumptionReportResponse, error)
	GetMeterReadings() (*client.MeterReadingsResponse, error)
	GetInverters() (*client.InvertersResponse, error)
}
