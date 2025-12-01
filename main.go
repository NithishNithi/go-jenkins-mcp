package main

import (
	"context"
	"os"

	"github.com/NithishNithi/go-jenkins-mcp/internal/config"
	"github.com/NithishNithi/go-jenkins-mcp/internal/mcp"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set up logging
	log := logrus.New()
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	// Create MCP server
	server, err := mcp.NewServer(cfg, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create MCP server")
	}

	// Start the server with stdio communication
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		log.WithError(err).Fatal("Server failed")
	}
}
