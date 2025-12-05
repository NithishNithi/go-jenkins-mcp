package jenkins

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NithishNithi/go-jenkins-mcp/internal/config"
	_ "github.com/leanovate/gopter" // Will be used for property-based testing
)

// JenkinsClient defines the interface for interacting with Jenkins API
type JenkinsClient interface {
	// Job operations
	ListJobs(ctx context.Context, folder string) ([]Job, error)
	GetJob(ctx context.Context, jobName string) (*JobDetails, error)

	// Build operations
	TriggerBuild(ctx context.Context, jobName string, params map[string]string) (*QueueItem, error)
	GetBuild(ctx context.Context, jobName string, buildNumber int) (*Build, error)
	GetLatestBuild(ctx context.Context, jobName string) (*Build, error)
	StopBuild(ctx context.Context, jobName string, buildNumber int) error

	// Log and artifact operations
	GetBuildLog(ctx context.Context, jobName string, buildNumber int) (string, error)
	ListArtifacts(ctx context.Context, jobName string, buildNumber int) ([]Artifact, error)
	GetArtifact(ctx context.Context, jobName string, buildNumber int, artifactPath string) ([]byte, error)

	// Queue operations
	GetQueue(ctx context.Context) ([]QueueItem, error)
	GetQueueItem(ctx context.Context, queueID int) (*QueueItem, error)
	CancelQueueItem(ctx context.Context, queueID int) error

	// Running builds operations
	GetRunningBuilds(ctx context.Context) ([]RunningBuild, error)

	// View operations
	ListViews(ctx context.Context) ([]View, error)
	GetView(ctx context.Context, viewName string) (*ViewDetails, error)
	CreateView(ctx context.Context, viewName string, viewType string) error
	GetNodes(ctx context.Context) ([]Node, error)
	GetPipelineScript(ctx context.Context, jobName string) (string, error)
}

// Client represents a Jenkins API client implementation
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	apiToken   string
	maxRetries int
	backoff    time.Duration
}

// retryTransport implements http.RoundTripper with retry logic and exponential backoff
type retryTransport struct {
	transport  http.RoundTripper
	maxRetries int
	backoff    time.Duration
}

// RoundTrip executes a single HTTP transaction with retry logic
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only retry idempotent operations (GET requests)
	if req.Method != http.MethodGet {
		return rt.transport.RoundTrip(req)
	}

	var lastErr error
	for attempt := 0; attempt <= rt.maxRetries; attempt++ {
		// Clone the request for retry attempts
		reqClone := req.Clone(req.Context())

		resp, err := rt.transport.RoundTrip(reqClone)

		// Success - return immediately
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		// Store the error
		lastErr = err
		if err == nil {
			// Server error - close the body and prepare for retry
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status code %d", resp.StatusCode)
		}

		// Don't sleep after the last attempt
		if attempt < rt.maxRetries {
			// Calculate exponential backoff: backoff * 2^attempt
			delay := time.Duration(float64(rt.backoff) * math.Pow(2, float64(attempt)))
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// NewClient creates a new Jenkins client with the provided configuration
func NewClient(cfg *config.Config) (JenkinsClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create base transport with TLS configuration
	transport, err := createTransport(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Wrap transport with retry logic
	retryTransport := &retryTransport{
		transport:  transport,
		maxRetries: cfg.MaxRetries,
		backoff:    cfg.RetryBackoff,
	}

	// Create HTTP client with timeout and custom transport
	httpClient := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: retryTransport,
	}

	client := &Client{
		baseURL:    cfg.JenkinsURL,
		httpClient: httpClient,
		username:   cfg.Username,
		password:   cfg.Password,
		apiToken:   cfg.APIToken,
		maxRetries: cfg.MaxRetries,
		backoff:    cfg.RetryBackoff,
	}

	return client, nil
}

// createTransport creates an HTTP transport with TLS configuration
func createTransport(cfg *config.Config) (*http.Transport, error) {
	// Start with default transport settings
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Configure TLS if needed
	if cfg.JenkinsURL[:5] == "https" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.TLSSkipVerify,
		}

		// Load custom CA certificate if provided
		if cfg.CACertPath != "" {
			caCert, err := os.ReadFile(cfg.CACertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}

			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}

			tlsConfig.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsConfig
	}

	return transport, nil
}

// addAuthentication adds authentication headers to the request
func (c *Client) addAuthentication(req *http.Request) {
	if c.apiToken != "" {
		// Use API token with basic auth (username + token as password)
		// Jenkins API tokens use basic auth with username and token
		req.SetBasicAuth(c.username, c.apiToken)
	} else if c.username != "" && c.password != "" {
		// Use basic authentication with username and password
		req.SetBasicAuth(c.username, c.password)
	}
}

// getCrumb fetches a CSRF crumb from Jenkins
func (c *Client) getCrumb(ctx context.Context) (string, string, error) {
	url := c.baseURL + "/crumbIssuer/api/json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create crumb request: %w", err)
	}

	// Add authentication
	c.addAuthentication(req)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("crumb request failed: %w", err)
	}
	defer resp.Body.Close()

	// If crumb issuer is not configured, return empty (no CSRF protection)
	if resp.StatusCode == http.StatusNotFound {
		return "", "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status code %d when fetching crumb", resp.StatusCode)
	}

	// Parse crumb response
	var crumbData struct {
		Crumb             string `json:"crumb"`
		CrumbRequestField string `json:"crumbRequestField"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read crumb response: %w", err)
	}

	if err := json.Unmarshal(body, &crumbData); err != nil {
		return "", "", fmt.Errorf("failed to parse crumb response: %w", err)
	}

	return crumbData.CrumbRequestField, crumbData.Crumb, nil
}

// doRequest executes an HTTP request with authentication and context
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	c.addAuthentication(req)

	// Set common headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// For POST requests, fetch and add CSRF crumb
	if method == http.MethodPost {
		crumbField, crumb, err := c.getCrumb(ctx)
		if err != nil {
			// Log the error but continue - some Jenkins instances don't have CSRF protection
			// In production, you might want to handle this differently
		} else if crumb != "" {
			// Add the crumb header
			req.Header.Set(crumbField, crumb)
		}
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Placeholder implementations for interface methods
// These will be implemented in subsequent tasks

func (c *Client) ListJobs(ctx context.Context, folder string) ([]Job, error) {
	// Build the API path
	path := "/api/json"
	if folder != "" {
		path = "/job/" + folder + path
	}

	// Add tree parameter to get specific job fields
	path += "?tree=jobs[name,url,description,buildable,inQueue,color]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("folder not found: %s", folder)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to list jobs")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Jobs []Job `json:"jobs"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Return empty list if no jobs (not an error)
	if result.Jobs == nil {
		return []Job{}, nil
	}

	return result.Jobs, nil
}

func (c *Client) GetJob(ctx context.Context, jobName string) (*JobDetails, error) {
	if jobName == "" {
		return nil, fmt.Errorf("job name cannot be empty")
	}

	// Build the API path with detailed tree parameter
	path := fmt.Sprintf("/job/%s/api/json", jobName)
	path += "?tree=name,url,description,buildable,inQueue,color,disabled,"
	path += "lastBuild[number,url],"
	path += "lastSuccessfulBuild[number,url],"
	path += "lastFailedBuild[number,url],"
	path += "property[parameterDefinitions[name,type,defaultParameterValue[value],description]]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("job not found: %s", jobName)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - Jenkins returns a complex structure for parameters
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// First parse into a raw structure to handle Jenkins' nested parameter format
	var rawResult struct {
		Name                string          `json:"name"`
		URL                 string          `json:"url"`
		Description         string          `json:"description"`
		Buildable           bool            `json:"buildable"`
		InQueue             bool            `json:"inQueue"`
		Color               string          `json:"color"`
		Disabled            bool            `json:"disabled"`
		LastBuild           *BuildReference `json:"lastBuild"`
		LastSuccessfulBuild *BuildReference `json:"lastSuccessfulBuild"`
		LastFailedBuild     *BuildReference `json:"lastFailedBuild"`
		Property            []struct {
			ParameterDefinitions []struct {
				Name                  string `json:"name"`
				Type                  string `json:"type"`
				Description           string `json:"description"`
				DefaultParameterValue struct {
					Value interface{} `json:"value"`
				} `json:"defaultParameterValue"`
			} `json:"parameterDefinitions"`
		} `json:"property"`
	}

	if err := json.Unmarshal(body, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Build JobDetails from raw result
	jobDetails := &JobDetails{
		Job: Job{
			Name:        rawResult.Name,
			URL:         rawResult.URL,
			Description: rawResult.Description,
			Buildable:   rawResult.Buildable,
			InQueue:     rawResult.InQueue,
			Color:       rawResult.Color,
		},
		LastBuild:           rawResult.LastBuild,
		LastSuccessfulBuild: rawResult.LastSuccessfulBuild,
		LastFailedBuild:     rawResult.LastFailedBuild,
		Disabled:            rawResult.Disabled,
		Parameters:          []JobParameter{},
	}

	// Extract parameters from the nested structure
	for _, prop := range rawResult.Property {
		for _, paramDef := range prop.ParameterDefinitions {
			param := JobParameter{
				Name:         paramDef.Name,
				Type:         paramDef.Type,
				Description:  paramDef.Description,
				DefaultValue: paramDef.DefaultParameterValue.Value,
			}
			jobDetails.Parameters = append(jobDetails.Parameters, param)
		}
	}

	return jobDetails, nil
}

func (c *Client) TriggerBuild(ctx context.Context, jobName string, params map[string]string) (*QueueItem, error) {
	if jobName == "" {
		return nil, fmt.Errorf("job name cannot be empty")
	}

	// First, get job details to validate parameters
	jobDetails, err := c.GetJob(ctx, jobName)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details for validation: %w", err)
	}

	// Validate parameters against job definition
	if len(params) > 0 {
		if err := c.validateParameters(jobDetails, params); err != nil {
			return nil, fmt.Errorf("parameter validation failed: %w", err)
		}
	}

	// Determine the endpoint based on whether parameters are provided
	var path string
	var body io.Reader

	if len(params) > 0 {
		// Use buildWithParameters endpoint with query parameters
		path = fmt.Sprintf("/job/%s/buildWithParameters", jobName)

		// Jenkins expects parameters as query parameters in the URL
		queryParams := url.Values{}
		for name, value := range params {
			queryParams.Add(name, value)
		}
		path = path + "?" + queryParams.Encode()
	} else {
		// Use simple build endpoint
		path = fmt.Sprintf("/job/%s/build", jobName)
	}

	// Make POST request
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger build: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("job not found: %s", jobName)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to trigger build for job %s", jobName)
	}

	// Handle redirects (302, 303, 307, 308) and success codes (201, 200)
	var location string
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		// Redirect response - get location header
		location = resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("redirect received but no Location header present (status: %d)", resp.StatusCode)
		}
	} else if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		// Success response - try to get location header
		location = resp.Header.Get("Location")
		if location == "" {
			// Some Jenkins instances don't return Location header
			// Try alternative: use the queue location pattern or return a delayed lookup
			// For now, we'll generate a queue location based on response
			location = c.generateQueueLocationFromResponse(jobName, resp)
			if location == "" {
				return nil, fmt.Errorf(
					"jenkins did not return a queue Location header. " +
						"This usually happens when:\n" +
						" - Authentication failed (MFA enforced)\n" +
						" - API token is invalid\n" +
						" - Reverse proxy removed Location header\n" +
						" - Build was not triggered at all\n\n" +
						"Tip: Check for MFA redirects or test with curl",
				)

			}
		}
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Extract queue item ID from location URL
	// Location format: http://jenkins.example.com/queue/item/{id}/
	queueID, err := c.parseQueueIDFromLocation(location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse queue ID from location: %w", err)
	}

	// Return queue item with the ID
	queueItem := &QueueItem{
		ID:      queueID,
		JobName: jobName,
	}

	// If parameters were provided, include them in the response
	if len(params) > 0 {
		queueItem.Parameters = params
	}

	return queueItem, nil
}

// validateParameters validates the provided parameters against the job definition
func (c *Client) validateParameters(jobDetails *JobDetails, params map[string]string) error {
	// Build a map of valid parameter names from job definition
	validParams := make(map[string]JobParameter)
	for _, param := range jobDetails.Parameters {
		validParams[param.Name] = param
	}

	// Check if all provided parameters are valid
	for paramName := range params {
		if _, exists := validParams[paramName]; !exists {
			return fmt.Errorf("invalid parameter: %s is not defined for this job", paramName)
		}
	}

	return nil
}

// parseQueueIDFromLocation extracts the queue item ID from the Location header
func (c *Client) parseQueueIDFromLocation(location string) (int, error) {
	// Location format: http://jenkins.example.com/queue/item/{id}/
	// We need to extract the ID from the URL

	// Find "/queue/item/" in the location
	queueItemPrefix := "/queue/item/"
	idx := bytes.Index([]byte(location), []byte(queueItemPrefix))
	if idx == -1 {
		return 0, fmt.Errorf("invalid location format: %s", location)
	}

	// Extract the part after "/queue/item/"
	idPart := location[idx+len(queueItemPrefix):]

	// Remove trailing slash if present
	idPartBytes := bytes.TrimSuffix([]byte(idPart), []byte("/"))
	idPart = string(idPartBytes)

	// Parse the ID
	var queueID int
	_, err := fmt.Sscanf(idPart, "%d", &queueID)
	if err != nil {
		return 0, fmt.Errorf("failed to parse queue ID from %s: %w", idPart, err)
	}

	return queueID, nil
}

// generateQueueLocationFromResponse attempts to generate a queue location when the Location header is missing
// This is a fallback for Jenkins instances that don't return the Location header in the response
func (c *Client) generateQueueLocationFromResponse(_ string, resp *http.Response) string {
	// If we can't get the Location header, we have a few options:
	// 1. Return a placeholder and let the caller fetch the queue via other means
	// 2. Attempt to parse the response body for queue information
	// 3. Return empty and handle gracefully in the calling code

	// For now, we'll read the response body to see if it contains queue information
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Reset the response body for potential re-reading
	// (Note: this only works for our current use case since we're at the end)

	// Try to extract queue ID from response body if it's a redirect HTML page
	bodyStr := string(bodyBytes)

	// Look for queue item URL in various formats
	// Pattern 1: Direct queue location in response
	queuePattern := regexp.MustCompile(`/queue/item/(\d+)/`)
	matches := queuePattern.FindStringSubmatch(bodyStr)
	if len(matches) > 1 {
		return matches[0] // Return the full path
	}

	// Pattern 2: Jenkins may return a JSON response with queue URL
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResp); err == nil {
		if queueURL, ok := jsonResp["_links"].(map[string]interface{})["self"].(map[string]interface{})["href"]; ok {
			return fmt.Sprintf("%v", queueURL)
		}
	}

	// If we still can't find it, return empty string to indicate failure
	return ""
}

func (c *Client) GetBuild(ctx context.Context, jobName string, buildNumber int) (*Build, error) {
	if jobName == "" {
		return nil, fmt.Errorf("job name cannot be empty")
	}
	if buildNumber <= 0 {
		return nil, fmt.Errorf("build number must be positive")
	}

	// Build the API path with tree parameter to get specific build fields
	path := fmt.Sprintf("/job/%s/%d/api/json", jobName, buildNumber)
	path += "?tree=number,url,result,building,duration,timestamp,executor,estimatedDuration"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get build: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("build not found: job=%s, build=%d", jobName, buildNumber)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access build for job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var build Build
	if err := json.Unmarshal(body, &build); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &build, nil
}

func (c *Client) GetLatestBuild(ctx context.Context, jobName string) (*Build, error) {
	if jobName == "" {
		return nil, fmt.Errorf("job name cannot be empty")
	}

	// Build the API path to get the lastBuild information
	path := fmt.Sprintf("/job/%s/api/json", jobName)
	path += "?tree=lastBuild[number,url,result,building,duration,timestamp,executor,estimatedDuration]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest build: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("job not found: %s", jobName)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		LastBuild *Build `json:"lastBuild"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if there is a last build
	if result.LastBuild == nil {
		return nil, fmt.Errorf("no builds found for job: %s", jobName)
	}

	return result.LastBuild, nil
}

func (c *Client) StopBuild(ctx context.Context, jobName string, buildNumber int) error {
	if jobName == "" {
		return fmt.Errorf("job name cannot be empty")
	}
	if buildNumber <= 0 {
		return fmt.Errorf("build number must be positive")
	}

	// First, check if the build exists and is running
	build, err := c.GetBuild(ctx, jobName, buildNumber)
	if err != nil {
		return fmt.Errorf("failed to get build status: %w", err)
	}

	// Check if the build is already completed
	if !build.Building {
		return fmt.Errorf("build is not running: job=%s, build=%d, status=%s", jobName, buildNumber, build.Result)
	}

	// Build the API path for stopping the build
	path := fmt.Sprintf("/job/%s/%d/stop", jobName, buildNumber)

	// Make POST request to stop the build
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("failed to stop build: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("build not found: job=%s, build=%d", jobName, buildNumber)
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("permission denied: insufficient permissions to stop build for job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Verify the build status has been updated to aborted
	// Give Jenkins a moment to update the status
	time.Sleep(500 * time.Millisecond)

	updatedBuild, err := c.GetBuild(ctx, jobName, buildNumber)
	if err != nil {
		return fmt.Errorf("failed to verify build status after stop: %w", err)
	}

	// Check if the build was successfully stopped
	if updatedBuild.Building {
		return fmt.Errorf("build is still running after stop request: job=%s, build=%d", jobName, buildNumber)
	}

	// Verify the result is ABORTED
	if updatedBuild.Result != "ABORTED" {
		return fmt.Errorf("build status is %s, expected ABORTED: job=%s, build=%d", updatedBuild.Result, jobName, buildNumber)
	}

	return nil
}

func (c *Client) GetBuildLog(ctx context.Context, jobName string, buildNumber int) (string, error) {
	return c.GetBuildLogWithLimit(ctx, jobName, buildNumber, 0)
}

// GetBuildLogWithLimit retrieves build console log with optional size limit
// If sizeLimit is 0, the entire log is retrieved
// If sizeLimit > 0, only the first sizeLimit bytes are retrieved
func (c *Client) GetBuildLogWithLimit(ctx context.Context, jobName string, buildNumber int, sizeLimit int64) (string, error) {
	if jobName == "" {
		return "", fmt.Errorf("job name cannot be empty")
	}
	if buildNumber <= 0 {
		return "", fmt.Errorf("build number must be positive")
	}
	if sizeLimit < 0 {
		return "", fmt.Errorf("size limit must be non-negative")
	}

	// Build the API path for console text
	path := fmt.Sprintf("/job/%s/%d/consoleText", jobName, buildNumber)

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get build log: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("build log not found: job=%s, build=%d", jobName, buildNumber)
	}
	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("permission denied: insufficient permissions to access build log for job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Read the log content with optional size limit
	var logBytes []byte
	if sizeLimit > 0 {
		// Use LimitReader to cap the amount of data read
		limitedReader := io.LimitReader(resp.Body, sizeLimit)
		logBytes, err = io.ReadAll(limitedReader)
	} else {
		// Read the entire log
		logBytes, err = io.ReadAll(resp.Body)
	}

	if err != nil {
		return "", fmt.Errorf("failed to read log content: %w", err)
	}

	// Return the log as a string, preserving all formatting including line breaks and ANSI codes
	// This handles both in-progress builds (partial content) and completed builds
	return string(logBytes), nil
}

func (c *Client) ListArtifacts(ctx context.Context, jobName string, buildNumber int) ([]Artifact, error) {
	if jobName == "" {
		return nil, fmt.Errorf("job name cannot be empty")
	}
	if buildNumber <= 0 {
		return nil, fmt.Errorf("build number must be positive")
	}

	// Build the API path with artifacts tree parameter
	path := fmt.Sprintf("/job/%s/%d/api/json", jobName, buildNumber)
	path += "?tree=artifacts[fileName,relativePath,size]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list artifacts: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("build not found: job=%s, build=%d", jobName, buildNumber)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access artifacts for job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Artifacts []Artifact `json:"artifacts"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Return empty list if no artifacts (not an error)
	if result.Artifacts == nil {
		return []Artifact{}, nil
	}

	return result.Artifacts, nil
}

func (c *Client) GetArtifact(ctx context.Context, jobName string, buildNumber int, artifactPath string) ([]byte, error) {
	if jobName == "" {
		return nil, fmt.Errorf("job name cannot be empty")
	}
	if buildNumber <= 0 {
		return nil, fmt.Errorf("build number must be positive")
	}
	if artifactPath == "" {
		return nil, fmt.Errorf("artifact path cannot be empty")
	}

	// Build the API path for artifact download
	path := fmt.Sprintf("/job/%s/%d/artifact/%s", jobName, buildNumber, artifactPath)

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("artifact not found: job=%s, build=%d, path=%s", jobName, buildNumber, artifactPath)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access artifact for job %s", jobName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Read artifact content efficiently
	// For large artifacts, this uses streaming internally via io.ReadAll
	artifactData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact content: %w", err)
	}

	return artifactData, nil
}

func (c *Client) GetQueue(ctx context.Context) ([]QueueItem, error) {
	// Build the API path with tree parameter to get specific queue fields
	path := "/queue/api/json"
	path += "?tree=items[id,task[name],why,blocked,buildable,stuck,inQueueSince,params]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access build queue")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Jenkins returns a structure with nested task information
	var rawResult struct {
		Items []struct {
			ID   int `json:"id"`
			Task struct {
				Name string `json:"name"`
			} `json:"task"`
			Why          string `json:"why"`
			Blocked      bool   `json:"blocked"`
			Buildable    bool   `json:"buildable"`
			Stuck        bool   `json:"stuck"`
			InQueueSince int64  `json:"inQueueSince"`
			Params       string `json:"params,omitempty"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Transform raw result into QueueItem slice
	queueItems := make([]QueueItem, 0, len(rawResult.Items))
	for _, item := range rawResult.Items {
		queueItem := QueueItem{
			ID:           item.ID,
			JobName:      item.Task.Name,
			Why:          item.Why,
			Blocked:      item.Blocked,
			Buildable:    item.Buildable,
			Stuck:        item.Stuck,
			InQueueSince: item.InQueueSince,
		}

		// Parse parameters if present
		// Jenkins may return params as a string that needs to be parsed
		if item.Params != "" {
			// For now, we'll store the raw params string
			// In a real implementation, this might need more sophisticated parsing
			queueItem.Parameters = make(map[string]string)
			// Note: Jenkins params format varies, this is a simplified approach
		}

		queueItems = append(queueItems, queueItem)
	}

	// Return empty list if no items (not an error)
	if len(queueItems) == 0 {
		return []QueueItem{}, nil
	}

	return queueItems, nil
}

func (c *Client) GetRunningBuilds(ctx context.Context) ([]RunningBuild, error) {
	// Get the list of all jobs first
	jobs, err := c.ListJobs(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	runningBuilds := []RunningBuild{}

	// For each job, check if it has a running build
	for _, job := range jobs {
		// Get job details to check the last build
		jobDetails, err := c.GetJob(ctx, job.Name)
		if err != nil {
			// Skip jobs we can't access
			continue
		}

		// Check if there's a last build
		if jobDetails.LastBuild == nil {
			continue
		}

		// Get the build details
		build, err := c.GetBuild(ctx, job.Name, jobDetails.LastBuild.Number)
		if err != nil {
			// Skip builds we can't access
			continue
		}

		// If the build is currently running, add it to the list
		if build.Building {
			runningBuild := RunningBuild{
				JobName:           job.Name,
				BuildNumber:       build.Number,
				URL:               build.URL,
				Timestamp:         build.Timestamp,
				EstimatedDuration: build.EstimatedDuration,
				Executor:          build.Executor,
			}
			runningBuilds = append(runningBuilds, runningBuild)
		}
	}

	return runningBuilds, nil
}

// GetQueueItem retrieves details about a specific queue item
func (c *Client) GetQueueItem(ctx context.Context, queueID int) (*QueueItem, error) {
	if queueID <= 0 {
		return nil, fmt.Errorf("queue ID must be positive")
	}

	// Build the API path
	path := fmt.Sprintf("/queue/item/%d/api/json", queueID)

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue item: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("queue item not found: %d", queueID)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access queue item %d", queueID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rawResult struct {
		ID   int `json:"id"`
		Task struct {
			Name string `json:"name"`
		} `json:"task"`
		Why          string `json:"why"`
		Blocked      bool   `json:"blocked"`
		Buildable    bool   `json:"buildable"`
		Stuck        bool   `json:"stuck"`
		InQueueSince int64  `json:"inQueueSince"`
		Params       string `json:"params,omitempty"`
	}

	if err := json.Unmarshal(body, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	queueItem := &QueueItem{
		ID:           rawResult.ID,
		JobName:      rawResult.Task.Name,
		Why:          rawResult.Why,
		Blocked:      rawResult.Blocked,
		Buildable:    rawResult.Buildable,
		Stuck:        rawResult.Stuck,
		InQueueSince: rawResult.InQueueSince,
	}

	return queueItem, nil
}

// CancelQueueItem cancels a queued build before it starts
func (c *Client) CancelQueueItem(ctx context.Context, queueID int) error {
	if queueID <= 0 {
		return fmt.Errorf("queue ID must be positive")
	}

	// Build the API path
	path := fmt.Sprintf("/queue/cancelItem?id=%d", queueID)

	// Make POST request
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel queue item: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("queue item not found: %d", queueID)
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("permission denied: insufficient permissions to cancel queue item %d", queueID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListViews retrieves all views
func (c *Client) ListViews(ctx context.Context) ([]View, error) {
	// Build the API path
	path := "/api/json?tree=views[name,url,description]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list views: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to list views")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Views []View `json:"views"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Return empty list if no views (not an error)
	if result.Views == nil {
		return []View{}, nil
	}

	return result.Views, nil
}

// GetView retrieves details about a specific view
func (c *Client) GetView(ctx context.Context, viewName string) (*ViewDetails, error) {
	if viewName == "" {
		return nil, fmt.Errorf("view name cannot be empty")
	}

	// Build the API path
	path := fmt.Sprintf("/view/%s/api/json?tree=name,url,description,jobs[name,url,description,buildable,inQueue,color]", viewName)

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get view: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("view not found: %s", viewName)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to access view %s", viewName)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var viewDetails ViewDetails
	if err := json.Unmarshal(body, &viewDetails); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &viewDetails, nil
}

// CreateView creates a new view
func (c *Client) CreateView(ctx context.Context, viewName string, viewType string) error {
	if viewName == "" {
		return fmt.Errorf("view name cannot be empty")
	}
	if viewType == "" {
		viewType = "hudson.model.ListView" // Default to list view
	}

	// Build the API path with URL-encoded view name
	path := fmt.Sprintf("/createView?name=%s", url.QueryEscape(viewName))

	// Create the view configuration XML
	viewConfig := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<%s>
  <name>%s</name>
  <description></description>
  <filterExecutors>false</filterExecutors>
  <filterQueue>false</filterQueue>
  <properties class="hudson.model.View$PropertyList"/>
  <jobNames>
    <comparator class="hudson.util.CaseInsensitiveComparator"/>
  </jobNames>
  <jobFilters/>
  <columns>
    <hudson.views.StatusColumn/>
    <hudson.views.WeatherColumn/>
    <hudson.views.JobColumn/>
    <hudson.views.LastSuccessColumn/>
    <hudson.views.LastFailureColumn/>
    <hudson.views.LastDurationColumn/>
    <hudson.views.BuildButtonColumn/>
  </columns>
</%s>`, viewType, viewName, viewType)

	// Make POST request with XML body
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader([]byte(viewConfig)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	c.addAuthentication(req)

	// Set headers for XML content
	req.Header.Set("Content-Type", "application/xml")

	// Add CSRF crumb for POST request
	crumbField, crumb, err := c.getCrumb(ctx)
	if err == nil && crumb != "" {
		req.Header.Set(crumbField, crumb)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("view already exists: %s", viewName)
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("permission denied: insufficient permissions to create view. Details: %s", string(body))
	}
	if resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad request: invalid view configuration. Details: %s", string(body))
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type Node struct {
	DisplayName        string `json:"displayName"`
	Offline            bool   `json:"offline"`
	TemporarilyOffline bool   `json:"temporarilyOffline"`
	NumExecutors       int    `json:"numExecutors"`
}

// GetNodes retrieves all Jenkins nodes
func (c *Client) GetNodes(ctx context.Context) ([]Node, error) {
	// Build API path (customize fields as needed)
	path := "/computer/api/json?tree=computer[displayName,offline,temporarilyOffline,numExecutors]"

	// Make GET request
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: insufficient permissions to get nodes")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Computer []Node `json:"computer"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Return empty list if no nodes (not an error)
	if result.Computer == nil {
		return []Node{}, nil
	}

	return result.Computer, nil
}
func (c *Client) GetPipelineScript(ctx context.Context, job string) (string, error) {
	path := fmt.Sprintf("/job/%s/config.xml", job)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch config.xml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("job not found: %s", job)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	configBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read config.xml: %w", err)
	}

	xml := string(configBytes)

	// ────────────────────────────────────────────────
	// Case 1: Inline Pipeline script (CpsFlowDefinition)
	// ────────────────────────────────────────────────
	if strings.Contains(xml, "org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition") {
		re := regexp.MustCompile(`<script>([\s\S]*?)</script>`)
		match := re.FindStringSubmatch(xml)
		if len(match) >= 2 {
			return match[1], nil
		}
		return "", fmt.Errorf("pipeline job found, but <script> block is empty or missing")
	}

	// ────────────────────────────────────────────────
	// Case 2: SCM Pipeline job (CpsScmFlowDefinition)
	// ────────────────────────────────────────────────
	if strings.Contains(xml, "org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition") {
		return "", fmt.Errorf("pipeline is defined in SCM (Git). Jenkins does not store the Jenkinsfile inline")
	}

	// ────────────────────────────────────────────────
	// Case 3: Multibranch / non-pipeline job
	// ────────────────────────────────────────────────
	return "", fmt.Errorf("job '%s' is not an inline pipeline job (no <script> block available)", job)
}
