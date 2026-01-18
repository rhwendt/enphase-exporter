package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var authLog = logrus.WithField("component", "auth")

const (
	// Enlighten API endpoints for token retrieval
	enlightenLoginURL = "https://enlighten.enphaseenergy.com/login/login.json"
	enlightenTokenURL = "https://entrez.enphaseenergy.com/tokens"

	// Session validity duration (Enphase sessions last ~10 minutes)
	sessionDuration = 10 * time.Minute
	// Refresh buffer - refresh session 1 minute before expiry
	refreshBuffer = 1 * time.Minute
)

// AuthResponse represents the response from /auth/check_jwt
type AuthResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	ManagerToken string `json:"manager_token"`
	IsConsumer bool   `json:"is_consumer"`
}

// EnlightenLoginResponse represents the response from Enlighten login
type EnlightenLoginResponse struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	IsConsumer bool  `json:"is_consumer"`
	ManagerToken string `json:"manager_token"`
}

// EnlightenTokenResponse represents the response from token endpoint
type EnlightenTokenResponse struct {
	Token      string `json:"token"`
	GenerationTime int64 `json:"generation_time"`
	ExpiresAt  int64  `json:"expires_at"`
}

// authenticate performs the full authentication flow.
func (c *Client) authenticate() error {
	var jwt string
	var err error

	// Use provided JWT or fetch from Enlighten
	if c.config.JWT != "" {
		authLog.Debug("Using provided JWT token")
		jwt = c.config.JWT
	} else {
		authLog.Debug("Fetching JWT from Enlighten")
		jwt, err = c.getEnlightenToken()
		if err != nil {
			return fmt.Errorf("failed to get Enlighten token: %w", err)
		}
	}

	// Validate JWT and get session cookie
	if err := c.validateJWT(jwt); err != nil {
		return fmt.Errorf("failed to validate JWT: %w", err)
	}

	return nil
}

// getEnlightenToken authenticates with Enlighten and retrieves a JWT token.
func (c *Client) getEnlightenToken() (string, error) {
	// Step 1: Login to Enlighten
	authLog.Debug("Logging into Enlighten")

	loginData := url.Values{}
	loginData.Set("user[email]", c.config.Username)
	loginData.Set("user[password]", c.config.Password)

	loginReq, err := http.NewRequest("POST", enlightenLoginURL, strings.NewReader(loginData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	loginResp, err := c.httpClient.Do(loginReq)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", loginResp.StatusCode, string(body))
	}

	var loginResponse EnlightenLoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&loginResponse); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	authLog.WithField("user_id", loginResponse.UserID).Debug("Enlighten login successful")

	// Step 2: Get JWT token for our serial number
	authLog.WithField("serial", c.config.Serial).Debug("Fetching JWT token")

	tokenURL := fmt.Sprintf("%s?serial_num=%s", enlightenTokenURL, c.config.Serial)
	tokenReq, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	tokenResp, err := c.httpClient.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tokenResp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", tokenResp.StatusCode, string(body))
	}

	// The token endpoint returns the JWT directly as text
	tokenBytes, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", fmt.Errorf("received empty token from Enlighten")
	}

	authLog.Debug("JWT token retrieved successfully")
	return token, nil
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
	// Extract session info from response for logging
	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		// Some firmware versions don't return JSON, that's OK
		authLog.Debug("Could not decode auth response (might be older firmware)")
	}

	// Set session expiry
	c.sessionID = "authenticated"
	c.sessionExp = time.Now().Add(sessionDuration)

	authLog.WithFields(logrus.Fields{
		"expires_at": c.sessionExp.Format(time.RFC3339),
	}).Info("Gateway session established")

	return nil
}

// refreshSession refreshes the session before it expires.
func (c *Client) refreshSession() error {
	authLog.Debug("Refreshing gateway session")
	return c.authenticate()
}

// isSessionValid checks if the current session is still valid.
func (c *Client) isSessionValid() bool {
	return c.sessionID != "" && time.Now().Add(refreshBuffer).Before(c.sessionExp)
}
