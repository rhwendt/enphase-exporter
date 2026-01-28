package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/rhwendt/enphase-exporter/internal/client"
	"github.com/rhwendt/enphase-exporter/internal/collector"
)

var (
	// Version information (set via ldflags at build time)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	log = logrus.New()

	// Global client reference for readiness checks
	envoyClient *client.Client
)

func main() {
	// Configure logging
	configureLogging()

	log.WithFields(logrus.Fields{
		"version": Version,
		"commit":  GitCommit,
		"built":   BuildDate,
	}).Info("Starting Enphase Exporter")

	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate required configuration
	if err := validateConfig(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Create Enphase client
	var err error
	envoyClient, err = client.New(client.Config{
		Address:  viper.GetString("envoy.address"),
		Serial:   viper.GetString("envoy.serial"),
		Username: viper.GetString("envoy.username"),
		Password: viper.GetString("envoy.password"),
		JWT:      viper.GetString("envoy.jwt"),
	})
	if err != nil {
		log.Fatalf("Failed to create Enphase client: %v", err)
	}

	log.WithFields(logrus.Fields{
		"address": viper.GetString("envoy.address"),
		"serial":  viper.GetString("envoy.serial"),
	}).Info("Configured Enphase gateway connection")

	// Authenticate on startup with retry logic
	// This ensures readiness probe passes and catches config issues early
	if err := authenticateWithRetry(envoyClient, 5, 5*time.Second); err != nil {
		log.Fatalf("Failed to authenticate with Enphase gateway: %v", err)
	}
	log.Info("Successfully authenticated with Enphase gateway")

	// Start proactive session refresh to prevent data gaps
	envoyClient.StartSessionRefresh()

	// Create and register collectors
	productionCollector := collector.NewProductionCollector(envoyClient)
	prometheus.MustRegister(productionCollector)

	metersCollector := collector.NewMetersCollector(envoyClient)
	prometheus.MustRegister(metersCollector)

	invertersCollector := collector.NewInvertersCollector(envoyClient)
	prometheus.MustRegister(invertersCollector)

	// Register build info metric
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "enphase_exporter_build_info",
			Help: "Build information for the enphase exporter",
		},
		[]string{"version", "commit", "built"},
	)
	prometheus.MustRegister(buildInfo)
	buildInfo.WithLabelValues(Version, GitCommit, BuildDate).Set(1)

	// Set up HTTP server
	port := viper.GetString("exporter.port")
	if port == "" {
		port = "9090"
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/", rootHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.WithField("signal", sig.String()).Info("Shutting down server")

	// Stop session refresh goroutine
	envoyClient.StopSessionRefresh()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited")
}

func configureLogging() {
	// Use JSON formatter for K8s log aggregation
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

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

	// JWT is required - generate at https://entrez.enphaseenergy.com
	jwt := viper.GetString("envoy.jwt")
	if jwt == "" {
		return errMissingConfig("ENVOY_JWT (generate at https://entrez.enphaseenergy.com)")
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

// authenticateWithRetry attempts to authenticate with exponential backoff.
// This handles transient network issues during startup.
func authenticateWithRetry(c *client.Client, maxRetries int, initialDelay time.Duration) error {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := c.Authenticate(); err != nil {
			lastErr = err
			log.WithFields(logrus.Fields{
				"attempt": attempt,
				"max":     maxRetries,
				"error":   err.Error(),
			}).Warn("Authentication attempt failed, retrying...")

			if attempt < maxRetries {
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
			continue
		}
		return nil // Success
	}

	return fmt.Errorf("authentication failed after %d attempts: %w", maxRetries, lastErr)
}

// healthHandler returns OK if the server is running (liveness probe).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyHandler returns OK only if the client has authenticated (readiness probe).
// If the session has expired, it triggers re-authentication to prevent the pod
// from becoming stuck in a not-ready state.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	if envoyClient == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Client not initialized"))
		return
	}

	// If already ready, return immediately
	if envoyClient.IsReady() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
		return
	}

	// Session may have expired - try to re-authenticate
	if err := envoyClient.Authenticate(); err != nil {
		log.WithError(err).Warn("Readiness check: re-authentication failed")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready: " + err.Error()))
		return
	}

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
<p>Version: ` + Version + `</p>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/health">Health</a> (liveness probe)</p>
<p><a href="/ready">Ready</a> (readiness probe)</p>
</body>
</html>`))
}
