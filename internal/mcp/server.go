package mcp

import (
	"context"
	"fmt"

	"github.com/NithishNithi/go-jenkins-mcp/internal/config"
	"github.com/NithishNithi/go-jenkins-mcp/internal/jenkins"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// Server represents the MCP server
type Server struct {
	config        *config.Config
	log           *logrus.Logger
	mcpServer     *mcp.Server
	jenkinsClient jenkins.JenkinsClient
}

// NewServer creates a new MCP server instance
func NewServer(cfg *config.Config, log *logrus.Logger) (*Server, error) {
	// Create Jenkins client
	jenkinsClient, err := jenkins.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jenkins client: %w", err)
	}

	// Create MCP server with implementation info
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "Go-Jenkins-MCPServer",
		Version: "1.0.0",
	}, nil)

	server := &Server{
		config:        cfg,
		log:           log,
		mcpServer:     mcpServer,
		jenkinsClient: jenkinsClient,
	}

	// Register all tools
	if err := server.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return server, nil
}

// Start starts the MCP server with stdio communication
func (s *Server) Start(ctx context.Context) error {
	s.log.WithFields(logrus.Fields{
		"transport":   "stdio",
		"jenkins_url": s.config.JenkinsURL,
	}).Info("Starting Jenkins MCP Server")

	// Start the server with stdio transport
	if err := s.mcpServer.Run(ctx, &mcp.StdioTransport{}); err != nil {
		s.log.WithError(err).Error("MCP server failed")
		return fmt.Errorf("MCP server failed: %w", err)
	}

	s.log.Info("MCP Server stopped gracefully")
	return nil
}

// registerTools registers all Jenkins tools with the MCP server
func (s *Server) registerTools() error {

	// ───────────────────────────────
	// JOBS
	// ───────────────────────────────
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_job",
		Description: "Get detailed information about a specific Jenkins job including configuration, parameters, and recent build history.",
	}, s.handleGetJob)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_list_jobs",
		Description: "List all accessible Jenkins jobs. Optionally filter by folder path.",
	}, s.handleListJobs)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_trigger_build",
		Description: "Trigger a new build for a Jenkins job. Supports parameterized builds.",
	}, s.handleTriggerBuild)

	// ───────────────────────────────
	// BUILDS
	// ───────────────────────────────
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_build",
		Description: "Get status and details of a specific build. If buildNumber is omitted, returns the latest build.",
	}, s.handleGetBuild)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_build_log",
		Description: "Retrieve the console output (log) for a specific build. Supports optional size limits for large logs.",
	}, s.handleGetBuildLog)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_running_builds",
		Description: "Get all currently running builds across all Jenkins jobs.",
	}, s.handleGetRunningBuilds)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_stop_build",
		Description: "Stop a running build. The build status will be updated to ABORTED.",
	}, s.handleStopBuild)

	// ───────────────────────────────
	// ARTIFACTS
	// ───────────────────────────────
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_artifact",
		Description: "Download a specific artifact from a build. Returns the artifact content.",
	}, s.handleGetArtifact)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_list_artifacts",
		Description: "List all artifacts produced by a specific build.",
	}, s.handleListArtifacts)

	// ───────────────────────────────
	// QUEUE
	// ───────────────────────────────
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_cancel_queue_item",
		Description: "Cancel a queued build before it starts.",
	}, s.handleCancelQueueItem)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_queue",
		Description: "Get the current Jenkins build queue showing all pending builds.",
	}, s.handleGetQueue)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_queue_item",
		Description: "Get details about a specific queue item by ID.",
	}, s.handleGetQueueItem)

	// ───────────────────────────────
	// VIEWS
	// ───────────────────────────────
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_create_view",
		Description: "Create a new Jenkins view.",
	}, s.handleCreateView)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_view",
		Description: "Get jobs in a specific Jenkins view.",
	}, s.handleGetView)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_list_views",
		Description: "List all Jenkins views.",
	}, s.handleListViews)

	// ───────────────────────────────
	// SERVER
	// ───────────────────────────────
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_server_health",
		Description: "Get the health status of the Jenkins server.",
	}, s.handleServerHealthStatus)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_list_nodes",
		Description: "List all Jenkins nodes in the network.",
	}, s.handleGetNodes)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "jenkins_get_pipeline_script",
		Description: "Retrieve the Jenkinsfile (pipeline script) of a pipeline job.",
	}, s.handleGetPipelineScript)

	s.log.WithFields(logrus.Fields{
		"tool_count": 20,
		"categories": []string{"jobs", "builds", "artifacts", "queue", "views", "server"},
	}).Info("Successfully registered all Jenkins tools")
	return nil
}
