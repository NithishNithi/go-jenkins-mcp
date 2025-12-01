package config

import (
	"os"
	"testing"
	"time"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://jenkins.example.com",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://jenkins.example.com",
			wantErr: false,
		},
		{
			name:    "valid URL with port",
			url:     "https://jenkins.example.com:8080",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "URL without scheme",
			url:     "jenkins.example.com",
			wantErr: true,
		},
		{
			name:    "URL with invalid scheme",
			url:     "ftp://jenkins.example.com",
			wantErr: true,
		},
		{
			name:    "URL without host",
			url:     "https://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{JenkinsURL: tt.url}
			err := cfg.ValidateURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "valid timeout",
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "minimum timeout",
			timeout: 1 * time.Second,
			wantErr: false,
		},
		{
			name:    "maximum timeout",
			timeout: 5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "zero timeout",
			timeout: 0,
			wantErr: true,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
			wantErr: true,
		},
		{
			name:    "timeout too small",
			timeout: 500 * time.Millisecond,
			wantErr: true,
		},
		{
			name:    "timeout too large",
			timeout: 10 * time.Minute,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Timeout: tt.timeout}
			err := cfg.ValidateTimeout()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with username/password",
			config: &Config{
				JenkinsURL:   "https://jenkins.example.com",
				Username:     "admin",
				Password:     "password",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with API token",
			config: &Config{
				JenkinsURL:   "https://jenkins.example.com",
				APIToken:     "token123",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: &Config{
				Username:     "admin",
				Password:     "password",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing authentication",
			config: &Config{
				JenkinsURL:   "https://jenkins.example.com",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				JenkinsURL:   "https://jenkins.example.com",
				Username:     "admin",
				Password:     "password",
				Timeout:      0,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: &Config{
				JenkinsURL:   "https://jenkins.example.com",
				Username:     "admin",
				Password:     "password",
				Timeout:      30 * time.Second,
				MaxRetries:   -1,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"JENKINS_URL":       os.Getenv("JENKINS_URL"),
		"JENKINS_API_TOKEN": os.Getenv("JENKINS_API_TOKEN"),
	}
	defer func() {
		// Restore original environment
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("JENKINS_URL", "https://test.jenkins.com")
	os.Setenv("JENKINS_API_TOKEN", "test-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.JenkinsURL != "https://test.jenkins.com" {
		t.Errorf("JenkinsURL = %v, want %v", cfg.JenkinsURL, "https://test.jenkins.com")
	}

	if cfg.APIToken != "test-token" {
		t.Errorf("APIToken = %v, want %v", cfg.APIToken, "test-token")
	}

	// Check defaults are applied
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want %v", cfg.MaxRetries, 3)
	}
}
