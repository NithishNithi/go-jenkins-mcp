package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds the configuration for the Jenkins MCP Server
type Config struct {
	JenkinsURL    string
	Username      string
	Password      string
	APIToken      string
	Timeout       time.Duration
	TLSSkipVerify bool
	CACertPath    string
	MaxRetries    int
	RetryBackoff  time.Duration
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Validate Jenkins URL
	if c.JenkinsURL == "" {
		return errors.New("jenkins URL is required")
	}

	if err := c.ValidateURL(); err != nil {
		return fmt.Errorf("invalid jenkins URL: %w", err)
	}

	// Validate authentication - either username/password or username/API token must be provided
	hasBasicAuth := c.Username != "" && c.Password != ""
	hasTokenAuth := c.Username != "" && c.APIToken != ""

	if !hasBasicAuth && !hasTokenAuth {
		return errors.New("authentication required: provide either username/password or username/API token")
	}

	// Ensure username is provided when using API token
	if c.APIToken != "" && c.Username == "" {
		return errors.New("username is required when using API token authentication")
	}

	// Validate timeout
	if err := c.ValidateTimeout(); err != nil {
		return err
	}

	// Validate retry settings
	if c.MaxRetries < 0 {
		return errors.New("max retries must be non-negative")
	}

	if c.RetryBackoff < 0 {
		return errors.New("retry backoff must be non-negative")
	}

	return nil
}

// ValidateURL validates the Jenkins URL format
func (c *Config) ValidateURL() error {
	if c.JenkinsURL == "" {
		return errors.New("URL cannot be empty")
	}

	parsedURL, err := url.Parse(c.JenkinsURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return errors.New("URL must include a scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return errors.New("URL must include a host")
	}

	return nil
}

// ValidateTimeout validates the timeout value
func (c *Config) ValidateTimeout() error {
	if c.Timeout < 0 {
		return errors.New("timeout must be non-negative")
	}

	if c.Timeout == 0 {
		return errors.New("timeout must be greater than zero")
	}

	// Set reasonable bounds - minimum 1 second, maximum 5 minutes
	minTimeout := 1 * time.Second
	maxTimeout := 5 * time.Minute

	if c.Timeout < minTimeout {
		return fmt.Errorf("timeout must be at least %v", minTimeout)
	}

	if c.Timeout > maxTimeout {
		return fmt.Errorf("timeout must not exceed %v", maxTimeout)
	}

	return nil
}

// Load loads configuration from environment variables or configuration file
// Configuration priority: defaults < config file < environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Load from configuration file if it exists
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.jenkins-mcp-server")
	v.AddConfigPath("/etc/jenkins-mcp-server")

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Bind environment variables
	bindEnvVariables(v)

	// Build config from viper
	cfg := &Config{
		JenkinsURL:    v.GetString("jenkins.url"),
		Username:      v.GetString("jenkins.username"),
		Password:      v.GetString("jenkins.password"),
		APIToken:      v.GetString("jenkins.apiToken"),
		Timeout:       v.GetDuration("jenkins.timeout"),
		TLSSkipVerify: v.GetBool("jenkins.tls.skipVerify"),
		CACertPath:    v.GetString("jenkins.tls.caCert"),
		MaxRetries:    v.GetInt("jenkins.retry.maxAttempts"),
		RetryBackoff:  v.GetDuration("jenkins.retry.backoff"),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("jenkins.timeout", 30*time.Second)
	v.SetDefault("jenkins.tls.skipVerify", false)
	v.SetDefault("jenkins.retry.maxAttempts", 3)
	v.SetDefault("jenkins.retry.backoff", 1*time.Second)
}

// bindEnvVariables binds environment variables to configuration keys
func bindEnvVariables(v *viper.Viper) {
	// Enable automatic environment variable binding
	v.AutomaticEnv()

	// Bind specific environment variables to config keys
	envBindings := map[string]string{
		"JENKINS_URL":             "jenkins.url",
		"JENKINS_USERNAME":        "jenkins.username",
		"JENKINS_PASSWORD":        "jenkins.password",
		"JENKINS_API_TOKEN":       "jenkins.apiToken",
		"JENKINS_TIMEOUT":         "jenkins.timeout",
		"JENKINS_TLS_SKIP_VERIFY": "jenkins.tls.skipVerify",
		"JENKINS_CA_CERT":         "jenkins.tls.caCert",
		"JENKINS_MAX_RETRIES":     "jenkins.retry.maxAttempts",
		"JENKINS_RETRY_BACKOFF":   "jenkins.retry.backoff",
	}

	for envVar, configKey := range envBindings {
		if val := os.Getenv(envVar); val != "" {
			v.Set(configKey, val)
		}
	}
}
