# Jenkins MCP Server

A Model Context Protocol (MCP) server implementation in Go that provides programmatic access to Jenkins CI/CD functionality. This server enables AI assistants and other MCP clients to interact with Jenkins instances, allowing them to query build status, trigger jobs, manage pipelines, and retrieve build artifacts through a standardized protocol interface.

## Features

- **Complete Jenkins API Coverage**: List jobs, get job details, trigger builds, monitor status, retrieve logs and artifacts
- **MCP Protocol Compliant**: Full implementation of the Model Context Protocol specification
- **Secure Authentication**: Username and API token authentication
- **Secure Connections**: TLS/SSL support with custom CA certificate configuration
- **Robust Error Handling**: Automatic retry with exponential backoff for transient failures
- **Configurable**: Environment variables or configuration file support
- **Production Ready**: Comprehensive error handling, logging, and timeout management

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Available Tools](#available-tools)
- [MCP Client Integration](#mcp-client-integration)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [License](#license)

## Installation

### Option 1: Install from Source

```bash
go install github.com/NithishNithi/go-jenkins-mcp/cmd/jenkins-mcp-server@latest
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/NithishNithi/go-jenkins-mcp.git
cd go-jenkins-mcp

# Build the binary
go build -o jenkins-mcp-server .

# Optionally, move to your PATH
sudo mv jenkins-mcp-server /usr/local/bin/
```

### Option 3: Download Pre-built Binary

Download the latest release from the [releases page](https://github.com/NithishNithi/go-jenkins-mcp/releases) and place it in your PATH.

### Requirements

- Go 1.23.6 or later (for building from source)
- Access to a Jenkins instance
- Jenkins username and API token

## Configuration

The Jenkins MCP Server can be configured using environment variables or a configuration file. Environment variables take precedence over configuration file settings.

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
```

### Configuration File

Create a `config.yaml` file:

```yaml
jenkins:
  url: https://jenkins.example.com
  username: your-username
  apiToken: your-api-token-here
  
  # Optional settings
  timeout: 30s
  
  tls:
    skipVerify: false
    caCert: /path/to/ca.crt
  
  retry:
    maxAttempts: 3
    backoff: 1s
```

Specify the config file when running:

```bash
jenkins-mcp-server --config /path/to/config.yaml
```

### Authentication

The server requires both username and API token for authentication:

1. Log in to Jenkins
2. Click your name in the top right
3. Click "Configure"
4. Under "API Token", click "Add new Token"
5. Copy the generated token and use it as `JENKINS_API_TOKEN`
6. Use your Jenkins username as `JENKINS_USERNAME`

## Usage

### Running the Server

The server communicates via stdio (standard input/output) as per the MCP specification:

```bash
# Using environment variables
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USERNAME=your-username
export JENKINS_API_TOKEN=your-token
jenkins-mcp-server

# Using configuration file
jenkins-mcp-server --config config.yaml
```

### Testing the Connection

You can test the server by sending MCP protocol messages via stdin. However, it's typically used through an MCP client like Claude Desktop.

## Available Tools

The Jenkins MCP Server exposes the following tools organized by category:

### Jobs

**jenkins_list_jobs** - List all accessible Jenkins jobs. Optionally filter by folder path.

**jenkins_get_job** - Get detailed information about a specific Jenkins job including configuration, parameters, and recent build history.

**jenkins_trigger_build** - Trigger a new build for a Jenkins job. Supports parameterized builds.

### Builds

**jenkins_get_build** - Get status and details of a specific build. If buildNumber is omitted, returns the latest build.

**jenkins_get_build_log** - Retrieve the console output (log) for a specific build. Supports optional size limits for large logs.

**jenkins_get_running_builds** - Get all currently running builds across all Jenkins jobs.

**jenkins_stop_build** - Stop a running build. The build status will be updated to ABORTED.

### Artifacts

**jenkins_list_artifacts** - List all artifacts produced by a specific build.

**jenkins_get_artifact** - Download a specific artifact from a build. Returns the artifact content.

### Queue

**jenkins_get_queue** - Get the current Jenkins build queue showing all pending builds.

**jenkins_get_queue_item** - Get details about a specific queue item by ID.

**jenkins_cancel_queue_item** - Cancel a queued build before it starts.

### Views

**jenkins_list_views** - List all Jenkins views.

**jenkins_get_view** - Get jobs in a specific Jenkins view.

**jenkins_create_view** - Create a new Jenkins view.

### Server & Nodes

**jenkins_server_health** - Get the health status of the Jenkins server.

**jenkins_list_nodes** - List all Jenkins nodes in the network.

**jenkins_get_pipeline_script** - Retrieve the Jenkinsfile (pipeline script) of a pipeline job.

## MCP Client Integration

### Claude Desktop

Add the Jenkins MCP Server to your Claude Desktop configuration:

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
        "ghcr.io/nithishnithi/jenkins-mcp-server:v1.0.2"
      ]
    }
  }
}
```

After updating the configuration, restart Claude Desktop. You can now ask Claude to interact with your Jenkins instance:

- "List all Jenkins jobs"
- "Trigger a build for the main-pipeline job"
- "Show me the latest build status for my-app"
- "Get the build log for build #42 of the deployment job"

### Other MCP Clients

Any MCP-compliant client can use this server. Configure the client to:
1. Execute the `jenkins-mcp-server` binary
2. Communicate via stdio
3. Pass required environment variables or config file path

## Troubleshooting

### Connection Issues

**Problem:** "Connection refused" or "Unable to connect to Jenkins"

**Solutions:**
- Verify `JENKINS_URL` is correct and accessible
- Check network connectivity to Jenkins instance
- Ensure Jenkins is running and responding
- Check firewall rules

### Authentication Failures

**Problem:** "Authentication failed" or "401 Unauthorized"

**Solutions:**
- Verify API token is valid and not expired
- Ensure username is correct
- Check that the user has necessary permissions
- Regenerate API token if needed

### TLS/SSL Errors

**Problem:** "x509: certificate signed by unknown authority"

**Solutions:**
- Set `JENKINS_TLS_SKIP_VERIFY=true` for testing (not recommended for production)
- Provide custom CA certificate via `JENKINS_CA_CERT` environment variable
- Ensure system CA certificates are up to date

### Permission Errors

**Problem:** "Permission denied" or "403 Forbidden"

**Solutions:**
- Verify the Jenkins user has appropriate permissions
- Check job-level permissions in Jenkins
- Ensure the user can perform the requested operation (trigger builds, stop builds, etc.)

### Timeout Issues

**Problem:** "Request timeout" or operations taking too long

**Solutions:**
- Increase `JENKINS_TIMEOUT` value
- Check Jenkins server performance
- Verify network latency to Jenkins instance
- Consider if Jenkins is under heavy load

### MCP Client Not Detecting Server

**Problem:** Claude Desktop or other MCP client doesn't show Jenkins tools

**Solutions:**
- Verify the binary path in MCP client configuration is correct
- Check that environment variables are properly set in the configuration
- Restart the MCP client after configuration changes
- Check MCP client logs for error messages
- Ensure the binary has execute permissions: `chmod +x /path/to/jenkins-mcp-server`

### Debugging

The server uses structured JSON logging with timestamps for better observability.

**Enable verbose logging:**

```bash
export LOG_LEVEL=debug
jenkins-mcp-server
```



## Development

### Project Structure

```
.
├── internal/
│   ├── config/                # Configuration management
│   ├── jenkins/               # Jenkins API client
│   └── mcp/                   # MCP server implementation
├── main.go                    # Main application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── README.md                  # Main documentation
├── QUICKSTART.md              # Quick start guide
├── DOCUMENTATION.md           # Documentation index
└── Dockerfile                 # Docker image definition
```

### Building

```bash
go build -o jenkins-mcp-server .
```

### Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: [GitHub Issues](https://github.com/NithishNithi/go-jenkins-mcp/issues)
- **Documentation**: [Full specification](./DOCUMENTATION.md)
- **MCP Protocol**: [Model Context Protocol](https://modelcontextprotocol.io/)

## Acknowledgments

- Built with the [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Inspired by the need for better AI-Jenkins integration
