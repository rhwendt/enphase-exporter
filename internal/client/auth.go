package client

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var authLog = logrus.WithField("component", "auth")

const (
	// Session validity duration (Enphase sessions last ~30 minutes)
	sessionDuration = 30 * time.Minute
	// Refresh buffer - refresh session 2 minutes before expiry
	refreshBuffer = 2 * time.Minute
)

// authenticate performs the authentication flow using JWT.
func (c *Client) authenticate() error {
	if c.config.JWT == "" {
		return fmt.Errorf("ENVOY_JWT is required. Generate a token at https://entrez.enphaseenergy.com")
	}

	authLog.Debug("Authenticating with JWT token")

	// Validate JWT and get session cookie from gateway
	if err := c.validateJWT(c.config.JWT); err != nil {
		return fmt.Errorf("failed to validate JWT: %w", err)
	}

	return nil
}

// validateJWT validates the JWT with the gateway and establishes a session.
func (c *Client) validateJWT(jwt string) error {
	authLog.Debug("Validating JWT with gateway")

	checkURL := fmt.Sprintf("%s/auth/check_jwt", c.config.Address)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create check_jwt request: %w", err)
	}

	// Set the JWT in the Authorization header
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("check_jwt request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("JWT validation failed with status %d: %s", resp.StatusCode, string(body))
	}

	// The session cookie is now stored in our cookie jar
	c.sessionID = "authenticated"
	c.sessionExp = time.Now().Add(sessionDuration)

	authLog.WithFields(logrus.Fields{
		"expires_at": c.sessionExp.Format(time.RFC3339),
	}).Info("Gateway session established")

	return nil
}

// isSessionValid checks if the current session is still valid.
func (c *Client) isSessionValid() bool {
	return c.sessionID != "" && time.Now().Add(refreshBuffer).Before(c.sessionExp)
}
