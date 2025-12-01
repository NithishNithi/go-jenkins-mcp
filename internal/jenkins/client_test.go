package jenkins

import (
	"context"
	"testing"
	"time"

	"github.com/NithishNithi/go-jenkins-mcp/internal/config"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config with username/password",
			cfg: &config.Config{
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
			cfg: &config.Config{
				JenkinsURL:   "https://jenkins.example.com",
				APIToken:     "test-token",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid config with TLS skip verify",
			cfg: &config.Config{
				JenkinsURL:    "https://jenkins.example.com",
				Username:      "admin",
				Password:      "password",
				Timeout:       30 * time.Second,
				TLSSkipVerify: true,
				MaxRetries:    3,
				RetryBackoff:  1 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "invalid config - missing URL",
			cfg: &config.Config{
				Username:     "admin",
				Password:     "password",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing authentication",
			cfg: &config.Config{
				JenkinsURL:   "https://jenkins.example.com",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClientImplementsInterface(t *testing.T) {
	// This test verifies that Client implements JenkinsClient interface
	cfg := &config.Config{
		JenkinsURL:   "https://jenkins.example.com",
		Username:     "admin",
		Password:     "password",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	// Verify client implements the interface
	var _ JenkinsClient = client
}

func TestClientAuthentication(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		checkFn  func(*Client) bool
		expected bool
	}{
		{
			name: "API token authentication",
			cfg: &config.Config{
				JenkinsURL:   "https://jenkins.example.com",
				APIToken:     "test-token",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			checkFn: func(c *Client) bool {
				return c.apiToken == "test-token" && c.username == "" && c.password == ""
			},
			expected: true,
		},
		{
			name: "username/password authentication",
			cfg: &config.Config{
				JenkinsURL:   "https://jenkins.example.com",
				Username:     "admin",
				Password:     "password",
				Timeout:      30 * time.Second,
				MaxRetries:   3,
				RetryBackoff: 1 * time.Second,
			},
			checkFn: func(c *Client) bool {
				return c.username == "admin" && c.password == "password" && c.apiToken == ""
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if err != nil {
				t.Fatalf("NewClient() failed: %v", err)
			}

			// Type assert to access internal fields for testing
			concreteClient, ok := client.(*Client)
			if !ok {
				t.Fatal("client is not of type *Client")
			}

			if tt.checkFn(concreteClient) != tt.expected {
				t.Errorf("authentication check failed")
			}
		})
	}
}

func TestClientTimeoutConfiguration(t *testing.T) {
	cfg := &config.Config{
		JenkinsURL:   "https://jenkins.example.com",
		Username:     "admin",
		Password:     "password",
		Timeout:      45 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	concreteClient, ok := client.(*Client)
	if !ok {
		t.Fatal("client is not of type *Client")
	}

	if concreteClient.httpClient.Timeout != 45*time.Second {
		t.Errorf("timeout = %v, want %v", concreteClient.httpClient.Timeout, 45*time.Second)
	}
}

func TestClientRetryConfiguration(t *testing.T) {
	cfg := &config.Config{
		JenkinsURL:   "https://jenkins.example.com",
		Username:     "admin",
		Password:     "password",
		Timeout:      30 * time.Second,
		MaxRetries:   5,
		RetryBackoff: 2 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	concreteClient, ok := client.(*Client)
	if !ok {
		t.Fatal("client is not of type *Client")
	}

	if concreteClient.maxRetries != 5 {
		t.Errorf("maxRetries = %v, want %v", concreteClient.maxRetries, 5)
	}

	if concreteClient.backoff != 2*time.Second {
		t.Errorf("backoff = %v, want %v", concreteClient.backoff, 2*time.Second)
	}
}

func TestPlaceholderMethods(t *testing.T) {
	cfg := &config.Config{
		JenkinsURL:   "https://jenkins.example.com",
		Username:     "admin",
		Password:     "password",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()

	// Test that implemented methods attempt to connect (will fail with network error in test)
	t.Run("ListJobs", func(t *testing.T) {
		_, err := client.ListJobs(ctx, "")
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("GetJob", func(t *testing.T) {
		_, err := client.GetJob(ctx, "test-job")
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("TriggerBuild", func(t *testing.T) {
		_, err := client.TriggerBuild(ctx, "test-job", nil)
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("GetBuild", func(t *testing.T) {
		_, err := client.GetBuild(ctx, "test-job", 1)
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("GetLatestBuild", func(t *testing.T) {
		_, err := client.GetLatestBuild(ctx, "test-job")
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("StopBuild", func(t *testing.T) {
		err := client.StopBuild(ctx, "test-job", 1)
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("GetBuildLog", func(t *testing.T) {
		_, err := client.GetBuildLog(ctx, "test-job", 1)
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("ListArtifacts", func(t *testing.T) {
		_, err := client.ListArtifacts(ctx, "test-job", 1)
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("GetArtifact", func(t *testing.T) {
		_, err := client.GetArtifact(ctx, "test-job", 1, "artifact.jar")
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})

	t.Run("GetQueue", func(t *testing.T) {
		_, err := client.GetQueue(ctx)
		if err == nil {
			t.Error("expected error when connecting to non-existent Jenkins instance")
		}
		// Should get a network error, not "not implemented"
	})
}

// Test artifact operations input validation
func TestListArtifactsValidation(t *testing.T) {
	cfg := &config.Config{
		JenkinsURL:   "https://jenkins.example.com",
		Username:     "admin",
		Password:     "password",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		jobName     string
		buildNumber int
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty job name",
			jobName:     "",
			buildNumber: 1,
			wantErr:     true,
			errContains: "job name cannot be empty",
		},
		{
			name:        "zero build number",
			jobName:     "test-job",
			buildNumber: 0,
			wantErr:     true,
			errContains: "build number must be positive",
		},
		{
			name:        "negative build number",
			jobName:     "test-job",
			buildNumber: -1,
			wantErr:     true,
			errContains: "build number must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.ListArtifacts(ctx, tt.jobName, tt.buildNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListArtifacts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("ListArtifacts() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestGetArtifactValidation(t *testing.T) {
	cfg := &config.Config{
		JenkinsURL:   "https://jenkins.example.com",
		Username:     "admin",
		Password:     "password",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		jobName      string
		buildNumber  int
		artifactPath string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "empty job name",
			jobName:      "",
			buildNumber:  1,
			artifactPath: "artifact.jar",
			wantErr:      true,
			errContains:  "job name cannot be empty",
		},
		{
			name:         "zero build number",
			jobName:      "test-job",
			buildNumber:  0,
			artifactPath: "artifact.jar",
			wantErr:      true,
			errContains:  "build number must be positive",
		},
		{
			name:         "negative build number",
			jobName:      "test-job",
			buildNumber:  -1,
			artifactPath: "artifact.jar",
			wantErr:      true,
			errContains:  "build number must be positive",
		},
		{
			name:         "empty artifact path",
			jobName:      "test-job",
			buildNumber:  1,
			artifactPath: "",
			wantErr:      true,
			errContains:  "artifact path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetArtifact(ctx, tt.jobName, tt.buildNumber, tt.artifactPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("GetArtifact() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
