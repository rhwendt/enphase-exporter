package main

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/rhwendt/enphase-exporter/internal/client"
	"github.com/rhwendt/enphase-exporter/internal/collector"
)

var log = logrus.New()

func main() {
	// Configure logging
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate required configuration
	if err := validateConfig(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Create Enphase client
	envoyClient, err := client.New(client.Config{
		Address:  viper.GetString("envoy.address"),
		Serial:   viper.GetString("envoy.serial"),
		Username: viper.GetString("envoy.username"),
		Password: viper.GetString("envoy.password"),
		JWT:      viper.GetString("envoy.jwt"),
	})
	if err != nil {
		log.Fatalf("Failed to create Enphase client: %v", err)
	}

	// Create and register collectors
	productionCollector := collector.NewProductionCollector(envoyClient)
	prometheus.MustRegister(productionCollector)

	metersCollector := collector.NewMetersCollector(envoyClient)
	prometheus.MustRegister(metersCollector)

	invertersCollector := collector.NewInvertersCollector(envoyClient)
	prometheus.MustRegister(invertersCollector)

	// Set up HTTP server
	port := viper.GetString("exporter.port")
	if port == "" {
		port = "9090"
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.HandleFunc("/", rootHandler)

	log.Infof("Starting Enphase Exporter on :%s", port)
	log.Infof("Metrics available at http://localhost:%s/metrics", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func loadConfig() error {
	// Bind environment variables
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Map environment variables to config keys
	viper.BindEnv("envoy.address", "ENVOY_ADDRESS")
	viper.BindEnv("envoy.serial", "ENVOY_SERIAL")
	viper.BindEnv("envoy.username", "ENVOY_USERNAME")
	viper.BindEnv("envoy.password", "ENVOY_PASSWORD")
	viper.BindEnv("envoy.jwt", "ENVOY_JWT")
	viper.BindEnv("exporter.port", "EXPORTER_PORT")
	viper.BindEnv("scrape.interval", "SCRAPE_INTERVAL")

	// Set defaults
	viper.SetDefault("exporter.port", "9090")
	viper.SetDefault("scrape.interval", 30)

	return nil
}

func validateConfig() error {
	address := viper.GetString("envoy.address")
	if address == "" {
		return errMissingConfig("ENVOY_ADDRESS")
	}

	serial := viper.GetString("envoy.serial")
	if serial == "" {
		return errMissingConfig("ENVOY_SERIAL")
	}

	// Need either JWT or username/password
	jwt := viper.GetString("envoy.jwt")
	username := viper.GetString("envoy.username")
	password := viper.GetString("envoy.password")

	if jwt == "" && (username == "" || password == "") {
		return errMissingConfig("ENVOY_JWT or ENVOY_USERNAME/ENVOY_PASSWORD")
	}

	return nil
}

type configError struct {
	field string
}

func (e *configError) Error() string {
	return "missing required configuration: " + e.field
}

func errMissingConfig(field string) error {
	return &configError{field: field}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Enphase Exporter</title></head>
<body>
<h1>Enphase Prometheus Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/health">Health</a></p>
<p><a href="/ready">Ready</a></p>
</body>
</html>`))
}

func init() {
	// Set log level from environment
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}
