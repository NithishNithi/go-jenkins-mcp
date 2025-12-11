# Jenkins MCP Server

A Model Context Protocol (MCP) server implementation in Go that provides programmatic access to Jenkins CI/CD functionality. This server enables AI assistants and other MCP clients to interact with Jenkins instances through a standardized protocol interface.

## Features

- **Complete Jenkins API Coverage**: List jobs, trigger builds, monitor status, retrieve logs and artifacts
- **MCP Protocol Compliant**: Full implementation of the Model Context Protocol specification
- **Secure Authentication**: Username and API token authentication with TLS/SSL support
- **Robust Error Handling**: Automatic retry with exponential backoff for transient failures
- **Production Ready**: Comprehensive error handling, logging, and timeout management
- **Multi-Instance Support**: Optional tool prefixing for running multiple Jenkins servers

## Installation

### Option 1: Install from Source

```bash
go install github.com/NithishNithi/go-jenkins-mcp/cmd/jenkins-mcp-server@latest
```

### Option 2: Build from Source

```bash
git clone https://github.com/NithishNithi/go-jenkins-mcp.git
cd go-jenkins-mcp
go build -o jenkins-mcp-server .
```

### Option 3: Download Pre-built Binary

Download the latest release from the [releases page](https://github.com/NithishNithi/go-jenkins-mcp/releases).

### Requirements

- Go 1.23.6 or later (for building from source)
- Access to a Jenkins instance
- Jenkins username and API token

## Configuration

### Environment Variables

```bash
# Required
JENKINS_URL=https://jenkins.example.com
JENKINS_USERNAME=your-username
JENKINS_API_TOKEN=your-api-token-here

# Optional
JENKINS_TIMEOUT=30s                    # Request timeout (default: 30s)
JENKINS_TLS_SKIP_VERIFY=false          # Skip TLS verification (default: false)
JENKINS_CA_CERT=/path/to/ca.crt        # Custom CA certificate path
JENKINS_MAX_RETRIES=3                  # Maximum retry attempts (default: 3)
JENKINS_RETRY_BACKOFF=1s               # Initial retry backoff (default: 1s)
JENKINS_TOOL_PREFIX=prod               # Tool name prefix for multi-instance setups
```

**Note:** `JENKINS_TOOL_PREFIX` allows you to run multiple Jenkins MCP servers simultaneously by prefixing tool names (e.g., `prod_jenkins_list_jobs`, `staging_jenkins_list_jobs`).

### Configuration File

Create a `config.yaml` file:

```yaml
jenkins:
  url: https://jenkins.example.com
  username: your-username
  apiToken: your-api-token-here
  toolPrefix: prod  # Optional prefix for tool names
  
  # Optional settings
  timeout: 30s
  tls:
    skipVerify: false
    caCert: /path/to/ca.crt
  retry:
    maxAttempts: 3
    backoff: 1s
```

Run with config file:

```bash
jenkins-mcp-server --config /path/to/config.yaml
```

### Getting Your Jenkins API Token

1. Log in to Jenkins
2. Click your name in the top right corner
3. Click "Configure"
4. Under "API Token", click "Add new Token"
5. Copy the generated token

## Usage

### Running the Server

```bash
# Using environment variables
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USERNAME=your-username
export JENKINS_API_TOKEN=your-token
jenkins-mcp-server

# Using configuration file
jenkins-mcp-server --config config.yaml
```

## Available Tools

### Jobs
- **jenkins_list_jobs** - List all accessible Jenkins jobs
- **jenkins_get_job** - Get detailed job information
- **jenkins_trigger_build** - Trigger a new build (supports parameters)

### Builds
- **jenkins_get_build** - Get build status and details
- **jenkins_get_build_log** - Retrieve console output
- **jenkins_get_running_builds** - Get all currently running builds
- **jenkins_stop_build** - Stop a running build

### Artifacts
- **jenkins_list_artifacts** - List build artifacts
- **jenkins_get_artifact** - Download specific artifacts

### Queue
- **jenkins_get_queue** - View the build queue
- **jenkins_get_queue_item** - Get queue item details
- **jenkins_cancel_queue_item** - Cancel queued builds

### Views
- **jenkins_list_views** - List all views
- **jenkins_get_view** - Get jobs in a view
- **jenkins_create_view** - Create a new view

### Server & Nodes
- **jenkins_server_health** - Check server health
- **jenkins_list_nodes** - List Jenkins nodes
- **jenkins_get_pipeline_script** - Retrieve Jenkinsfile content

## MCP Client Integration

### Claude Desktop

Add to your Claude Desktop configuration file:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "jenkins": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "JENKINS_URL=https://jenkins.example.com",
        "-e", "JENKINS_USERNAME=your-username",
        "-e", "JENKINS_API_TOKEN=your-api-token",
        "-e", "JENKINS_TOOL_PREFIX=prod",
        "ghcr.io/nithishnithi/jenkins-mcp-server:latest"
      ]
    }
  }
}
```

Restart Claude Desktop, then ask:
- "List all Jenkins jobs"
- "Trigger a build for the main-pipeline job"
- "Show me the latest build status for my-app"
- "Get the build log for build #42"

## Troubleshooting

### Connection Issues
- Verify `JENKINS_URL` is correct and accessible
- Check network connectivity and firewall rules
- Ensure Jenkins is running

### Authentication Failures
- Verify API token is valid
- Check username is correct
- Regenerate API token if needed

### TLS/SSL Errors
- Set `JENKINS_TLS_SKIP_VERIFY=true` for testing only
- Provide custom CA certificate via `JENKINS_CA_CERT`
- Update system CA certificates

### Permission Errors
- Verify Jenkins user has appropriate permissions
- Check job-level permissions in Jenkins

### Timeout Issues
- Increase `JENKINS_TIMEOUT` value
- Check Jenkins server performance
- Verify network latency

### MCP Client Not Detecting Server
- Verify binary path in configuration
- Check environment variables are properly set
- Restart MCP client after changes
- Ensure binary has execute permissions: `chmod +x jenkins-mcp-server`

### Enable Debug Logging

```bash
export LOG_LEVEL=debug
jenkins-mcp-server
```

## Development

### Project Structure

```
.
├── internal/
│   ├── config/      # Configuration management
│   ├── jenkins/     # Jenkins API client
│   └── mcp/         # MCP server implementation
├── main.go          # Application entry point
├── go.mod           # Go module definition
├── Dockerfile       # Docker image definition
└── README.md        # This file
```

### Building

```bash
go build -o jenkins-mcp-server .
```

### Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Add tests for your changes
4. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/NithishNithi/go-jenkins-mcp/issues)
- **MCP Protocol**: [Model Context Protocol](https://modelcontextprotocol.io/)

## Acknowledgments

Built with the [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)