package client

// API endpoint paths for Enphase IQ Gateway
const (
	// Public endpoints (no auth required)
	EndpointInfo = "/info.xml"

	// Production endpoints
	EndpointProduction = "/production.json"
	EndpointProductionDetails = "/production.json?details=1"

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
