package client

// ProductionResponse represents the response from /production.json
type ProductionResponse struct {
	Production  []ProductionDevice `json:"production"`
	Consumption []ProductionDevice `json:"consumption"`
	Storage     []StorageDevice    `json:"storage"`
}

// ProductionDevice represents a production device in the response.
type ProductionDevice struct {
	Type             string  `json:"type"`
	MeasurementType  string  `json:"measurementType,omitempty"` // For consumption: "total-consumption" or "net-consumption"
	ActiveCount      int     `json:"activeCount"`
	ReadingTime      int64   `json:"readingTime"`
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	WhToday          float64 `json:"whToday"`
	WhLastSevenDays  float64 `json:"whLastSevenDays"`
	VahLifetime      float64 `json:"vahLifetime"`
	RmsCurrent       float64 `json:"rmsCurrent"`
	RmsVoltage       float64 `json:"rmsVoltage"`
	ReactPwr         float64 `json:"reactPwr"`
	ApprntPwr        float64 `json:"apprntPwr"`
	PwrFactor        float64 `json:"pwrFactor"`
	WhLastUpdate     int64   `json:"whLastUpdate"`
	Lines            []Line  `json:"lines,omitempty"`
}

// Line represents per-phase data.
type Line struct {
	WNow            float64 `json:"wNow"`
	WhLifetime      float64 `json:"whLifetime"`
	WhToday         float64 `json:"whToday"`
	WhLastSevenDays float64 `json:"whLastSevenDays"`
	VahLifetime     float64 `json:"vahLifetime"`
	RmsCurrent      float64 `json:"rmsCurrent"`
	RmsVoltage      float64 `json:"rmsVoltage"`
	ReactPwr        float64 `json:"reactPwr"`
	ApprntPwr       float64 `json:"apprntPwr"`
	PwrFactor       float64 `json:"pwrFactor"`
}

// StorageDevice represents a storage device (battery).
type StorageDevice struct {
	Type        string  `json:"type"`
	ActiveCount int     `json:"activeCount"`
	ReadingTime int64   `json:"readingTime"`
	WNow        float64 `json:"wNow"`
	WhNow       float64 `json:"whNow"`
	State       string  `json:"state"`
}

// MeterReadingsResponse represents the response from /ivp/meters/readings
type MeterReadingsResponse []MeterReading

// MeterReading represents a single meter reading.
type MeterReading struct {
	Eid               int64          `json:"eid"`
	Timestamp         int64          `json:"timestamp"`
	ActEnergyDlvd     float64        `json:"actEnergyDlvd"`
	ActEnergyRcvd     float64        `json:"actEnergyRcvd"`
	ApparentEnergy    float64        `json:"apparentEnergy"`
	ReactEnergyLagg   float64        `json:"reactEnergyLagg"`
	ReactEnergyLead   float64        `json:"reactEnergyLead"`
	InstantaneousDemand float64      `json:"instantaneousDemand"`
	ActivePower       float64        `json:"activePower"`
	ApparentPower     float64        `json:"apparentPower"`
	ReactivePower     float64        `json:"reactivePower"`
	PwrFactor         float64        `json:"pwrFactor"`
	Voltage           float64        `json:"voltage"`
	Current           float64        `json:"current"`
	Freq              float64        `json:"freq"`
	Channels          []MeterChannel `json:"channels"`
}

// MeterChannel represents a phase/channel in the meter reading.
type MeterChannel struct {
	Eid               int64   `json:"eid"`
	Timestamp         int64   `json:"timestamp"`
	ActEnergyDlvd     float64 `json:"actEnergyDlvd"`
	ActEnergyRcvd     float64 `json:"actEnergyRcvd"`
	ApparentEnergy    float64 `json:"apparentEnergy"`
	ReactEnergyLagg   float64 `json:"reactEnergyLagg"`
	ReactEnergyLead   float64 `json:"reactEnergyLead"`
	InstantaneousDemand float64 `json:"instantaneousDemand"`
	ActivePower       float64 `json:"activePower"`
	ApparentPower     float64 `json:"apparentPower"`
	ReactivePower     float64 `json:"reactivePower"`
	PwrFactor         float64 `json:"pwrFactor"`
	Voltage           float64 `json:"voltage"`
	Current           float64 `json:"current"`
	Freq              float64 `json:"freq"`
}

// InvertersResponse represents the response from /api/v1/production/inverters
type InvertersResponse []Inverter

// Inverter represents a single inverter (panel).
type Inverter struct {
	SerialNumber string `json:"serialNumber"`
	LastReportDate int64 `json:"lastReportDate"`
	DevType      int    `json:"devType"`
	LastReportWatts int `json:"lastReportWatts"`
	MaxReportWatts int  `json:"maxReportWatts"`
}
