package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

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
}

// New creates a new Enphase client.
func New(config Config) (*Client, error) {
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

// GetProduction fetches production data from the gateway.
func (c *Client) GetProduction() (*ProductionResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	// TODO: Implement API call
	return nil, fmt.Errorf("not implemented")
}

// GetMeterReadings fetches meter readings from the gateway.
func (c *Client) GetMeterReadings() (*MeterReadingsResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	// TODO: Implement API call
	return nil, fmt.Errorf("not implemented")
}

// GetInverters fetches inverter data from the gateway.
func (c *Client) GetInverters() (*InvertersResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	// TODO: Implement API call
	return nil, fmt.Errorf("not implemented")
}

// ensureAuthenticated ensures we have a valid session.
func (c *Client) ensureAuthenticated() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if session is still valid (with 1 minute buffer)
	if c.sessionID != "" && time.Now().Add(time.Minute).Before(c.sessionExp) {
		return nil
	}

	// Need to authenticate
	return c.authenticate()
}

// authenticate performs the authentication flow.
func (c *Client) authenticate() error {
	// TODO: Implement authentication
	// 1. If JWT provided, use it directly
	// 2. Otherwise, get JWT from Enlighten using username/password
	// 3. Call /auth/check_jwt to validate and get session cookie
	return nil
}
