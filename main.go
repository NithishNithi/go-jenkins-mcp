package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NithishNithi/go-jenkins-mcp/internal/config"
	"github.com/NithishNithi/go-jenkins-mcp/internal/mcp"
	"github.com/sirupsen/logrus"
)

// OrderedJSONFormatter formats logs with a specific field order
type OrderedJSONFormatter struct {
	TimestampFormat string
}

// Format renders a single log entry with ordered fields
func (f *OrderedJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{})

	// Copy all fields
	for k, v := range entry.Data {
		data[k] = v
	}

	// Build ordered output
	var buf bytes.Buffer
	buf.WriteString("{")

	// 1. Timestamp (always first)
	timestamp := entry.Time.Format(f.TimestampFormat)
	buf.WriteString(fmt.Sprintf(`"timestamp":"%s"`, timestamp))

	// 2. Level (always second)
	buf.WriteString(fmt.Sprintf(`,"level":"%s"`, entry.Level.String()))

	// 3. Tool (if present, third)
	if tool, ok := data["tool"]; ok {
		toolJSON, _ := json.Marshal(tool)
		buf.WriteString(fmt.Sprintf(`,"tool":%s`, toolJSON))
		delete(data, "tool")
	}

	// 4. Message (always after tool)
	msgJSON, _ := json.Marshal(entry.Message)
	buf.WriteString(fmt.Sprintf(`,"message":%s`, msgJSON))

	// 5. Add remaining fields in alphabetical order
	if len(data) > 0 {
		remainingJSON, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		// Remove the outer braces and add to our buffer
		remainingStr := string(remainingJSON)
		if len(remainingStr) > 2 {
			buf.WriteString(",")
			buf.WriteString(remainingStr[1 : len(remainingStr)-1])
		}
	}

	buf.WriteString("}\n")
	return buf.Bytes(), nil
}

func main() {
	// Set up logging with enhanced formatting
	log := logrus.New()
	log.SetOutput(os.Stderr)

	// Set log level from environment or default to Info
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn", "warning":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Use custom ordered JSON formatter for structured logging
	log.SetFormatter(&OrderedJSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	log.WithFields(logrus.Fields{
		"version": "1.0.0",
		"service": "jenkins-mcp-server",
	}).Info("Initializing Jenkins MCP Server")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	log.WithFields(logrus.Fields{
		"jenkins_url": cfg.JenkinsURL,
		"username":    cfg.Username,
		"timeout":     cfg.Timeout,
	}).Info("Configuration loaded successfully")

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
