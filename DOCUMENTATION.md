# Documentation Index

Complete documentation for the Jenkins MCP Server.

## Getting Started

**New to Jenkins MCP Server?** Start here:

1. **[QUICKSTART.md](QUICKSTART.md)** - Get up and running in 5 minutes
2. **[README.md](README.md)** - Complete project overview and documentation

## Documentation by Topic

### Installation
- [README.md - Installation](README.md#installation)
- [QUICKSTART.md - Step 2](QUICKSTART.md#step-2-choose-your-installation-method)

### Configuration
- [README.md - Configuration](README.md#configuration)
- [QUICKSTART.md - Step 1](QUICKSTART.md#step-1-get-jenkins-api-token)

### Usage
- [README.md - Usage](README.md#usage)
- [README.md - Available Tools](README.md#available-tools)
- [QUICKSTART.md - Step 4](QUICKSTART.md#step-4-test-it-out)

### MCP Client Integration
- [README.md - MCP Client Integration](README.md#mcp-client-integration)
- [QUICKSTART.md - Step 3](QUICKSTART.md#step-3-integrate-with-claude-desktop)

### Troubleshooting
- [README.md - Troubleshooting](README.md#troubleshooting)
- [QUICKSTART.md - Common Issues](QUICKSTART.md#common-issues)

### Development
- [README.md - Development](README.md#development)

## Documentation Files

| File | Description |
|------|-------------|
| [README.md](README.md) | Main project documentation |
| [QUICKSTART.md](QUICKSTART.md) | 5-minute quick start guide |
| [DOCUMENTATION.md](DOCUMENTATION.md) | This file - documentation index |
| [Dockerfile](Dockerfile) | Docker image definition |
| [docker-compose.yaml](docker-compose.yaml) | Docker Compose for MCP server |

## Quick Reference

### Common Tasks

**Install the server:**
```bash
go install github.com/NithishNithi/go-jenkins-mcp/cmd/jenkins-mcp-server@latest
```

**Build from source:**
```bash
go build -o jenkins-mcp-server .
```

**Run with Docker:**
```bash
docker run -i \
  -e JENKINS_URL= \
  -e JENKINS_USERNAME= \
  -e JENKINS_API_TOKEN= \
  ghcr.io/nithishnithi/jenkins-mcp-server:v1.0.2
```

**Configure for Claude Desktop:**
```json
{
  "mcpServers": {
    "jenkins": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "JENKINS_URL=",
        "-e", "JENKINS_USERNAME=",
        "-e", "JENKINS_API_TOKEN=",
        "ghcr.io/nithishnithi/jenkins-mcp-server:v1.0.2"
      ]
    }
  }
}
```

## Support

- **Issues**: [GitHub Issues](https://github.com/NithishNithi/go-jenkins-mcp/issues)
- **Specifications**: [.kiro/specs/jenkins-mcp-server/](.kiro/specs/jenkins-mcp-server/)
- **MCP Protocol**: [Model Context Protocol](https://modelcontextprotocol.io/)
- **Jenkins API**: [Jenkins Remote Access API](https://www.jenkins.io/doc/book/using/remote-access-api/)

## Contributing

See [README.md - Contributing](README.md#contributing) for contribution guidelines.

## License

See [README.md - License](README.md#license) for license information.

---

**Need help?** Start with [QUICKSTART.md](QUICKSTART.md) or check the [Troubleshooting](#troubleshooting) sections.
