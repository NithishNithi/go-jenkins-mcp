# Documentation Index

Complete documentation for the Jenkins MCP Server.

## Getting Started

**New to Jenkins MCP Server?** Start here:

1. **[QUICKSTART.md](QUICKSTART.md)** - Get up and running in 5 minutes
2. **[README.md](README.md)** - Complete project overview and reference

## Documentation by Topic

### Installation
- [Installation Methods](README.md#installation) - Install from source, pre-built binaries, or Docker
- [Quick Start Installation](QUICKSTART.md#step-2-choose-your-installation-method) - Docker and Docker Compose setup

### Configuration
- [Environment Variables](README.md#environment-variables) - All available configuration options
- [Configuration File](README.md#configuration-file) - YAML-based configuration
- [Authentication Setup](README.md#getting-your-jenkins-api-token) - How to obtain Jenkins API tokens
- [Tool Prefix Configuration](README.md#environment-variables) - Setup for multi-instance deployments

### Usage
- [Running the Server](README.md#running-the-server) - Command-line usage
- [Available Tools](README.md#available-tools) - Complete list of Jenkins operations
- [Testing Commands](QUICKSTART.md#step-4-test-it-out) - Example Claude queries

### MCP Client Integration
- [Claude Desktop Setup](README.md#claude-desktop) - Configuration for Claude Desktop
- [Quick Integration Guide](QUICKSTART.md#step-3-integrate-with-claude-desktop) - Step-by-step Claude setup
- [Other MCP Clients](README.md#mcp-client-integration) - Integration with other MCP-compatible clients

### Troubleshooting
- [Common Issues](README.md#troubleshooting) - Solutions for frequent problems
- [Quick Troubleshooting](QUICKSTART.md#troubleshooting) - Fast fixes for common errors
- [Debug Logging](README.md#enable-debug-logging) - Enable verbose logging

### Development
- [Project Structure](README.md#project-structure) - Codebase organization
- [Building from Source](README.md#building) - Build instructions
- [Contributing Guidelines](README.md#contributing) - How to contribute

## Documentation Files

| File | Description |
|------|-------------|
| [README.md](README.md) | Main project documentation and reference |
| [QUICKSTART.md](QUICKSTART.md) | 5-minute quick start guide |
| [DOCUMENTATION.md](DOCUMENTATION.md) | This file - documentation index |
| [Dockerfile](Dockerfile) | Docker image definition |
| [docker-compose.yaml](docker-compose.yaml) | Docker Compose configuration |

## Quick Reference

### Installation Commands

**Install from source:**
```bash
go install github.com/NithishNithi/go-jenkins-mcp/cmd/jenkins-mcp-server@latest
```

**Build locally:**
```bash
git clone https://github.com/NithishNithi/go-jenkins-mcp.git
cd go-jenkins-mcp
go build -o jenkins-mcp-server .
```

**Run with Docker:**
```bash
docker run -i \
  -e JENKINS_URL=https://your-jenkins.com \
  -e JENKINS_USERNAME=your-username \
  -e JENKINS_API_TOKEN=your-token \
  -e JENKINS_TOOL_PREFIX=prod \
  ghcr.io/nithishnithi/jenkins-mcp-server:latest
```

### Configuration Examples

**Environment variables:**
```bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USERNAME=your-username
export JENKINS_API_TOKEN=your-api-token
export JENKINS_TOOL_PREFIX=prod
export JENKINS_TIMEOUT=30s
export JENKINS_TLS_SKIP_VERIFY=false
```

**Claude Desktop configuration:**
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

### Available Tools Summary

**Jobs:** `list_jobs`, `get_job`, `trigger_build`  
**Builds:** `get_build`, `get_build_log`, `get_running_builds`, `stop_build`  
**Artifacts:** `list_artifacts`, `get_artifact`  
**Queue:** `get_queue`, `get_queue_item`, `cancel_queue_item`  
**Views:** `list_views`, `get_view`, `create_view`  
**Server:** `server_health`, `list_nodes`, `get_pipeline_script`

See [README.md - Available Tools](README.md#available-tools) for detailed descriptions.

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JENKINS_URL` | Yes | - | Jenkins instance URL |
| `JENKINS_USERNAME` | Yes | - | Jenkins username |
| `JENKINS_API_TOKEN` | Yes | - | Jenkins API token |
| `JENKINS_TOOL_PREFIX` | No | - | Prefix for tool names (multi-instance support) |
| `JENKINS_TIMEOUT` | No | `30s` | Request timeout |
| `JENKINS_TLS_SKIP_VERIFY` | No | `false` | Skip TLS certificate verification |
| `JENKINS_CA_CERT` | No | - | Custom CA certificate path |
| `JENKINS_MAX_RETRIES` | No | `3` | Maximum retry attempts |
| `JENKINS_RETRY_BACKOFF` | No | `1s` | Initial retry backoff duration |

## Common Use Cases

### Multi-Instance Setup
Run multiple Jenkins servers simultaneously using `JENKINS_TOOL_PREFIX`:

```json
{
  "mcpServers": {
    "jenkins-prod": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "JENKINS_URL=https://prod-jenkins.com",
        "-e", "JENKINS_USERNAME=user",
        "-e", "JENKINS_API_TOKEN=token",
        "-e", "JENKINS_TOOL_PREFIX=prod",
        "ghcr.io/nithishnithi/jenkins-mcp-server:latest"
      ]
    },
    "jenkins-staging": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "JENKINS_URL=https://staging-jenkins.com",
        "-e", "JENKINS_USERNAME=user",
        "-e", "JENKINS_API_TOKEN=token",
        "-e", "JENKINS_TOOL_PREFIX=staging",
        "ghcr.io/nithishnithi/jenkins-mcp-server:latest"
      ]
    }
  }
}
```

### Self-Signed Certificates
For development environments with self-signed certificates:

```bash
docker run -i \
  -e JENKINS_URL=https://jenkins.local \
  -e JENKINS_USERNAME=admin \
  -e JENKINS_API_TOKEN=token \
  -e JENKINS_TLS_SKIP_VERIFY=true \
  ghcr.io/nithishnithi/jenkins-mcp-server:latest
```

**Warning:** Never use `JENKINS_TLS_SKIP_VERIFY=true` in production.

### Custom Timeouts
For slow Jenkins instances or large builds:

```bash
docker run -i \
  -e JENKINS_URL=https://jenkins.example.com \
  -e JENKINS_USERNAME=user \
  -e JENKINS_API_TOKEN=token \
  -e JENKINS_TIMEOUT=60s \
  -e JENKINS_MAX_RETRIES=5 \
  ghcr.io/nithishnithi/jenkins-mcp-server:latest
```

## Support Resources

- **GitHub Issues**: [Report bugs or request features](https://github.com/NithishNithi/go-jenkins-mcp/issues)
- **MCP Protocol**: [Model Context Protocol Documentation](https://modelcontextprotocol.io/)
- **Jenkins API**: [Jenkins Remote Access API](https://www.jenkins.io/doc/book/using/remote-access-api/)

## Contributing

Contributions are welcome! See [README.md - Contributing](README.md#contributing) for guidelines on:
- Forking the repository
- Creating feature branches
- Adding tests
- Submitting pull requests

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Need help?** Start with [QUICKSTART.md](QUICKSTART.md) for a 5-minute setup guide.