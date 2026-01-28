package client

// MeterReport represents a single report from /ivp/meters/reports/*
type MeterReport struct {
	CreatedAt  int64             `json:"createdAt"`
	ReportType string            `json:"reportType"` // "production", "total-consumption", "net-consumption"
	Cumulative MeterReportData   `json:"cumulative"`
	Lines      []MeterReportData `json:"lines,omitempty"`
}

// MeterReportData contains power/energy data from a meter report.
type MeterReportData struct {
	CurrW       float64 `json:"currW"`
	ActPower    float64 `json:"actPower"`
	ApprntPwr   float64 `json:"apprntPwr"`
	ReactPwr    float64 `json:"reactPwr"`
	WhDlvdCum   float64 `json:"whDlvdCum"`
	WhRcvdCum   float64 `json:"whRcvdCum"`
	VarhLagCum  float64 `json:"varhLagCum"`
	VarhLeadCum float64 `json:"varhLeadCum"`
	VahCum      float64 `json:"vahCum"`
	RmsVoltage  float64 `json:"rmsVoltage"`
	RmsCurrent  float64 `json:"rmsCurrent"`
	PwrFactor   float64 `json:"pwrFactor"`
	FreqHz      float64 `json:"freqHz"`
}

// ProductionReportResponse is a single MeterReport from /ivp/meters/reports/production
type ProductionReportResponse = MeterReport

// ConsumptionReportResponse is an array of MeterReports from /ivp/meters/reports/consumption
// Contains "total-consumption" and "net-consumption" entries.
type ConsumptionReportResponse = []MeterReport

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
