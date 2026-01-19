package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var clientLog = logrus.WithField("component", "client")

// Config holds the configuration for the Enphase client.
type Config struct {
	Address  string
	Serial   string
	Username string
	Password string
	JWT      string
}

// Client is an HTTP client for the Enphase IQ Gateway.
type Client struct {
	config     Config
	httpClient *http.Client
	sessionID  string
	sessionExp time.Time
	mu         sync.RWMutex
	ready      bool
}

// New creates a new Enphase client.
func New(config Config) (*Client, error) {
	// Validate required config
	if config.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if config.Serial == "" {
		return nil, fmt.Errorf("serial is required")
	}

	// Create cookie jar for session management
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Create HTTP client with TLS skip verify (local gateway uses self-signed cert)
	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client := &Client{
		config:     config,
		httpClient: httpClient,
	}

	return client, nil
}

// Address returns the gateway address.
func (c *Client) Address() string {
	return c.config.Address
}

// IsReady returns true if the client has successfully authenticated.
// Used for K8s readiness probes.
func (c *Client) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready && c.isSessionValid()
}

// Authenticate performs initial authentication with the gateway.
// Call this on startup to ensure the exporter is ready before serving requests.
func (c *Client) Authenticate() error {
	return c.ensureAuthenticated()
}

// GetProduction fetches production data from the gateway.
func (c *Client) GetProduction() (*ProductionResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	url := c.config.Address + EndpointProductionDetails
	resp, err := c.doRequest("GET", url)
	if err != nil {
		return nil, fmt.Errorf("production request failed: %w", err)
	}
	defer resp.Body.Close()

	var result ProductionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode production response: %w", err)
	}

	return &result, nil
}

// GetMeterReadings fetches meter readings from the gateway.
func (c *Client) GetMeterReadings() (*MeterReadingsResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	url := c.config.Address + EndpointMeterReadings
	resp, err := c.doRequest("GET", url)
	if err != nil {
		return nil, fmt.Errorf("meter readings request failed: %w", err)
	}
	defer resp.Body.Close()

	var result MeterReadingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode meter readings response: %w", err)
	}

	return &result, nil
}

// GetInverters fetches inverter data from the gateway.
func (c *Client) GetInverters() (*InvertersResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	url := c.config.Address + EndpointInverters
	resp, err := c.doRequest("GET", url)
	if err != nil {
		return nil, fmt.Errorf("inverters request failed: %w", err)
	}
	defer resp.Body.Close()

	var result InvertersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode inverters response: %w", err)
	}

	return &result, nil
}

// doRequest performs an HTTP request with proper error handling.
func (c *Client) doRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("request returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// ensureAuthenticated ensures we have a valid session.
func (c *Client) ensureAuthenticated() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if session is still valid
	if c.isSessionValid() {
		return nil
	}

	// Need to authenticate
	if err := c.authenticate(); err != nil {
		c.ready = false
		return err
	}

	c.ready = true
	return nil
}
