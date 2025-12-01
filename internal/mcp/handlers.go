package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NithishNithi/go-jenkins-mcp/internal/jenkins"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerHealthArgs defines the input parameters for jenkins_server_health
type ServerHealthArgs struct {
	// No parameters needed for server health check
}

// ListJobsArgs defines the input parameters for jenkins_list_jobs
type ListJobsArgs struct {
	Folder string `json:"folder,omitempty" jsonschema_description:"Optional folder path to list jobs from a specific folder"`
}

// handleListJobs handles the jenkins_list_jobs tool call
func (s *Server) handleListJobs(ctx context.Context, request *mcp.CallToolRequest, args ListJobsArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	jobs, err := s.jenkinsClient.ListJobs(ctx, args.Folder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetJobArgs defines the input parameters for jenkins_get_job
type GetJobArgs struct {
	JobName string `json:"jobName" jsonschema_description:"Name of the Jenkins job"`
}

// handleGetJob handles the jenkins_get_job tool call
func (s *Server) handleGetJob(ctx context.Context, request *mcp.CallToolRequest, args GetJobArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	jobDetails, err := s.jenkinsClient.GetJob(ctx, args.JobName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get job details: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(jobDetails, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// TriggerBuildArgs defines the input parameters for jenkins_trigger_build
type TriggerBuildArgs struct {
	JobName    string            `json:"jobName" jsonschema_description:"Name of the Jenkins job to trigger"`
	Parameters map[string]string `json:"parameters,omitempty" jsonschema_description:"Optional build parameters as key-value pairs"`
}

// handleTriggerBuild handles the jenkins_trigger_build tool call
func (s *Server) handleTriggerBuild(ctx context.Context, request *mcp.CallToolRequest, args TriggerBuildArgs) (*mcp.CallToolResult, any, error) {
	// First, get job details to check if it has parameters
	jobDetails, err := s.jenkinsClient.GetJob(ctx, args.JobName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get job details: %w", err)
	}

	// Check if the job has parameters defined
	if len(jobDetails.Parameters) > 0 {
		// Job has parameters - check if user provided them
		if len(args.Parameters) == 0 {
			// Parameters are required but not provided
			// Build a helpful message for the AI to ask the user
			var paramList []string
			for _, param := range jobDetails.Parameters {
				defaultVal := ""
				if param.DefaultValue != nil {
					defaultVal = fmt.Sprintf(" (default: %v)", param.DefaultValue)
				}
				paramDesc := param.Description
				if paramDesc != "" {
					paramDesc = " - " + paramDesc
				}
				paramList = append(paramList, fmt.Sprintf("  • %s (%s)%s%s", param.Name, param.Type, defaultVal, paramDesc))
			}

			warningMsg := fmt.Sprintf(`⚠️  PARAMETERS REQUIRED

The job '%s' requires parameters before it can be triggered.

Required Parameters:
%s

🤖 AI Action Required:
Please ask the user to provide values for these parameters before triggering the build.

Example:
"What values would you like to use for these parameters?"

Once the user provides the parameters, call this tool again with the parameters included.`,
				args.JobName,
				fmt.Sprintf("%s", paramList))

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: warningMsg},
				},
			}, nil, nil
		}

		// Parameters provided - validate that all required parameters are present
		providedParams := make(map[string]bool)
		for key := range args.Parameters {
			providedParams[key] = true
		}

		var missingParams []string
		for _, param := range jobDetails.Parameters {
			if !providedParams[param.Name] {
				// Check if parameter has a default value
				if param.DefaultValue == nil || param.DefaultValue == "" {
					missingParams = append(missingParams, param.Name)
				}
			}
		}

		if len(missingParams) > 0 {
			warningMsg := fmt.Sprintf(`⚠️  MISSING REQUIRED PARAMETERS
The following required parameters are missing:
%s

🤖 AI Action Required:
Please ask the user to provide values for these missing parameters.`,
				fmt.Sprintf("  • %s", missingParams))

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: warningMsg},
				},
			}, nil, nil
		}
	}

	// All validations passed - trigger the build
	queueItem, err := s.jenkinsClient.TriggerBuild(ctx, args.JobName, args.Parameters)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to trigger build: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(queueItem, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	successMsg := fmt.Sprintf("✅ Build triggered successfully!\n\n%s", string(result))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: successMsg},
		},
	}, nil, nil
}

// GetBuildArgs defines the input parameters for jenkins_get_build
type GetBuildArgs struct {
	JobName     string `json:"jobName" jsonschema_description:"Name of the Jenkins job"`
	BuildNumber *int   `json:"buildNumber,omitempty" jsonschema_description:"Build number (optional, omit to get latest build)"`
}

// handleGetBuild handles the jenkins_get_build tool call
func (s *Server) handleGetBuild(ctx context.Context, request *mcp.CallToolRequest, args GetBuildArgs) (*mcp.CallToolResult, any, error) {
	var build interface{}
	var err error

	if args.BuildNumber != nil {
		// Get specific build
		build, err = s.jenkinsClient.GetBuild(ctx, args.JobName, *args.BuildNumber)
	} else {
		// Get latest build
		build, err = s.jenkinsClient.GetLatestBuild(ctx, args.JobName)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get build: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(build, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetBuildLogArgs defines the input parameters for jenkins_get_build_log
type GetBuildLogArgs struct {
	JobName     string `json:"jobName" jsonschema_description:"Name of the Jenkins job"`
	BuildNumber int    `json:"buildNumber" jsonschema_description:"Build number"`
	SizeLimit   *int64 `json:"sizeLimit,omitempty" jsonschema_description:"Optional maximum size in bytes (0 for unlimited)"`
}

// handleGetBuildLog handles the jenkins_get_build_log tool call
func (s *Server) handleGetBuildLog(ctx context.Context, request *mcp.CallToolRequest, args GetBuildLogArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	var log string
	var err error

	if args.SizeLimit != nil && *args.SizeLimit > 0 {
		// Use the internal method with size limit
		if client, ok := s.jenkinsClient.(*jenkins.Client); ok {
			log, err = client.GetBuildLogWithLimit(ctx, args.JobName, args.BuildNumber, *args.SizeLimit)
		} else {
			// Fallback to regular method
			log, err = s.jenkinsClient.GetBuildLog(ctx, args.JobName, args.BuildNumber)
		}
	} else {
		log, err = s.jenkinsClient.GetBuildLog(ctx, args.JobName, args.BuildNumber)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get build log: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: log},
		},
	}, nil, nil
}

// ListArtifactsArgs defines the input parameters for jenkins_list_artifacts
type ListArtifactsArgs struct {
	JobName     string `json:"jobName" jsonschema_description:"Name of the Jenkins job"`
	BuildNumber int    `json:"buildNumber" jsonschema_description:"Build number"`
}

// handleListArtifacts handles the jenkins_list_artifacts tool call
func (s *Server) handleListArtifacts(ctx context.Context, request *mcp.CallToolRequest, args ListArtifactsArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	artifacts, err := s.jenkinsClient.ListArtifacts(ctx, args.JobName, args.BuildNumber)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list artifacts: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetArtifactArgs defines the input parameters for jenkins_get_artifact
type GetArtifactArgs struct {
	JobName      string `json:"jobName" jsonschema_description:"Name of the Jenkins job"`
	BuildNumber  int    `json:"buildNumber" jsonschema_description:"Build number"`
	ArtifactPath string `json:"artifactPath" jsonschema_description:"Relative path of the artifact"`
}

// handleGetArtifact handles the jenkins_get_artifact tool call
func (s *Server) handleGetArtifact(ctx context.Context, request *mcp.CallToolRequest, args GetArtifactArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	artifactData, err := s.jenkinsClient.GetArtifact(ctx, args.JobName, args.BuildNumber, args.ArtifactPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	// Encode artifact as base64 for safe transport
	encoded := base64.StdEncoding.EncodeToString(artifactData)

	// Return with metadata
	response := map[string]interface{}{
		"artifactPath": args.ArtifactPath,
		"size":         len(artifactData),
		"encoding":     "base64",
		"content":      encoded,
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetQueueArgs defines the input parameters for jenkins_get_queue (no parameters needed)
type GetQueueArgs struct{}

// handleGetQueue handles the jenkins_get_queue tool call
func (s *Server) handleGetQueue(ctx context.Context, request *mcp.CallToolRequest, args GetQueueArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	queueItems, err := s.jenkinsClient.GetQueue(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get queue: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(queueItems, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// StopBuildArgs defines the input parameters for jenkins_stop_build
type StopBuildArgs struct {
	JobName     string `json:"jobName" jsonschema_description:"Name of the Jenkins job"`
	BuildNumber int    `json:"buildNumber" jsonschema_description:"Build number to stop"`
}

// handleStopBuild handles the jenkins_stop_build tool call
func (s *Server) handleStopBuild(ctx context.Context, request *mcp.CallToolRequest, args StopBuildArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	err := s.jenkinsClient.StopBuild(ctx, args.JobName, args.BuildNumber)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stop build: %w", err)
	}

	// Return confirmation
	response := map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("Build %d for job %s has been stopped", args.BuildNumber, args.JobName),
		"jobName":     args.JobName,
		"buildNumber": args.BuildNumber,
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// handleServerHealthStatus handles the jenkins_server_health tool call
func (s *Server) handleServerHealthStatus(ctx context.Context, request *mcp.CallToolRequest, args ServerHealthArgs) (*mcp.CallToolResult, any, error) {
	url := s.config.JenkinsURL
	healthURL := fmt.Sprintf("%s/health", url)

	// Send an HTTP GET request to the /health endpoint
	resp, err := http.Get(healthURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get server health status from %s: %w", healthURL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var healthResponse struct {
		Status bool `json:"status"`
	}

	if err := json.Unmarshal(body, &healthResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !healthResponse.Status {
		return nil, nil, fmt.Errorf("server health status is false")
	}

	result, err := json.MarshalIndent(healthResponse, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetRunningBuildsArgs defines the input parameters for jenkins_get_running_builds
type GetRunningBuildsArgs struct {
	// No parameters needed - returns all running builds
}

// handleGetRunningBuilds handles the jenkins_get_running_builds tool call
func (s *Server) handleGetRunningBuilds(ctx context.Context, request *mcp.CallToolRequest, args GetRunningBuildsArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	runningBuilds, err := s.jenkinsClient.GetRunningBuilds(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get running builds: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(runningBuilds, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetQueueItemArgs defines the input parameters for jenkins_get_queue_item
type GetQueueItemArgs struct {
	QueueID int `json:"queueId" jsonschema_description:"Queue item ID"`
}

// handleGetQueueItem handles the jenkins_get_queue_item tool call
func (s *Server) handleGetQueueItem(ctx context.Context, request *mcp.CallToolRequest, args GetQueueItemArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	queueItem, err := s.jenkinsClient.GetQueueItem(ctx, args.QueueID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get queue item: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(queueItem, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// CancelQueueItemArgs defines the input parameters for jenkins_cancel_queue_item
type CancelQueueItemArgs struct {
	QueueID int `json:"queueId" jsonschema_description:"Queue item ID to cancel"`
}

// handleCancelQueueItem handles the jenkins_cancel_queue_item tool call
func (s *Server) handleCancelQueueItem(ctx context.Context, request *mcp.CallToolRequest, args CancelQueueItemArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	err := s.jenkinsClient.CancelQueueItem(ctx, args.QueueID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to cancel queue item: %w", err)
	}

	// Return success message
	successMsg := fmt.Sprintf("Successfully cancelled queue item %d", args.QueueID)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: successMsg},
		},
	}, nil, nil
}

// ListViewsArgs defines the input parameters for jenkins_list_views
type ListViewsArgs struct {
	// No parameters needed
}

// handleListViews handles the jenkins_list_views tool call
func (s *Server) handleListViews(ctx context.Context, request *mcp.CallToolRequest, args ListViewsArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	views, err := s.jenkinsClient.ListViews(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list views: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(views, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// GetViewArgs defines the input parameters for jenkins_get_view
type GetViewArgs struct {
	ViewName string `json:"viewName" jsonschema_description:"Name of the view"`
}

// handleGetView handles the jenkins_get_view tool call
func (s *Server) handleGetView(ctx context.Context, request *mcp.CallToolRequest, args GetViewArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	viewDetails, err := s.jenkinsClient.GetView(ctx, args.ViewName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get view: %w", err)
	}

	// Convert to JSON for response
	result, err := json.MarshalIndent(viewDetails, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

// CreateViewArgs defines the input parameters for jenkins_create_view
type CreateViewArgs struct {
	ViewName string `json:"viewName" jsonschema_description:"Name of the new view"`
	ViewType string `json:"viewType,omitempty" jsonschema_description:"Type of view (default: hudson.model.ListView)"`
}

// handleCreateView handles the jenkins_create_view tool call
func (s *Server) handleCreateView(ctx context.Context, request *mcp.CallToolRequest, args CreateViewArgs) (*mcp.CallToolResult, any, error) {
	// Call Jenkins client
	err := s.jenkinsClient.CreateView(ctx, args.ViewName, args.ViewType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create view: %w", err)
	}

	// Return success message
	successMsg := fmt.Sprintf("Successfully created view '%s'", args.ViewName)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: successMsg},
		},
	}, nil, nil
}
