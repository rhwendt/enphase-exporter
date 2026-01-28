package client

// API endpoint paths for Enphase IQ Gateway
const (
	// Public endpoints (no auth required)
	EndpointInfo = "/info.xml"

	// Production/consumption report endpoints (fast ~0.5s each)
	EndpointProductionReport  = "/ivp/meters/reports/production"
	EndpointConsumptionReport = "/ivp/meters/reports/consumption"

	// Meter endpoints
	EndpointMeterReadings = "/ivp/meters/readings"
	EndpointMeters = "/ivp/meters"

	// Inverter endpoints
	EndpointInverters = "/api/v1/production/inverters"

	// Inventory endpoints
	EndpointInventory = "/inventory.json"

	// Authentication endpoints
	EndpointAuthCheckJWT = "/auth/check_jwt"

	// Stream endpoints (WebSocket - future use)
	EndpointStreamMeter = "/stream/meter"
)
