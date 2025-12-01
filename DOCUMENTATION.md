# Documentation Index

Complete documentation for the Jenkins MCP Server.

## Getting Started

**New to Jenkins MCP Server?** Start here:

1. **[QUICKSTART.md](QUICKSTART.md)** - Get up and running in 5 minutes
2. **[README.md](README.md)** - Complete project overview and documentation

## Configuration

Learn how to configure the server:

- **[examples/config.yaml](examples/config.yaml)** - Example configuration file
- **[examples/.env.example](examples/.env.example)** - Example environment variables
- **[examples/CONFIGURATION.md](examples/CONFIGURATION.md)** - Complete configuration reference

## Deployment

Choose your deployment method:

- **[examples/DEPLOYMENT.md](examples/DEPLOYMENT.md)** - All deployment methods (local, Docker, Kubernetes, MCP clients)
- **[examples/DOCKER_DEPLOYMENT.md](examples/DOCKER_DEPLOYMENT.md)** - Docker-specific deployment guide
- **[Dockerfile](Dockerfile)** - Docker image definition
- **[docker-compose.yaml](docker-compose.yaml)** - Docker Compose configuration

## Testing

Set up a test environment:

- **[examples/TESTING.md](examples/TESTING.md)** - Local Jenkins setup for testing
- **[docker-compose-jenkins.yaml](docker-compose-jenkins.yaml)** - Run Jenkins locally

## Documentation by Topic

### Installation
- [README.md - Installation](README.md#installation)
- [QUICKSTART.md - Step 2](QUICKSTART.md#step-2-choose-your-installation-method)

### Configuration
- [README.md - Configuration](README.md#configuration)
- [examples/CONFIGURATION.md](examples/CONFIGURATION.md)
- [QUICKSTART.md - Step 1](QUICKSTART.md#step-1-get-jenkins-api-token)

### Usage
- [README.md - Usage](README.md#usage)
- [README.md - Available Tools](README.md#available-tools)
- [QUICKSTART.md - Step 4](QUICKSTART.md#step-4-test-it-out)

### MCP Client Integration
- [README.md - MCP Client Integration](README.md#mcp-client-integration)
- [examples/DEPLOYMENT.md - MCP Client Integration](examples/DEPLOYMENT.md#mcp-client-integration)
- [QUICKSTART.md - Step 3](QUICKSTART.md#step-3-integrate-with-claude-desktop)

### Docker Deployment
- [examples/DOCKER_DEPLOYMENT.md](examples/DOCKER_DEPLOYMENT.md)
- [examples/DEPLOYMENT.md - Docker Deployment](examples/DEPLOYMENT.md#docker-deployment)

### Kubernetes Deployment
- [examples/DEPLOYMENT.md - Kubernetes Deployment](examples/DEPLOYMENT.md#kubernetes-deployment)

### Troubleshooting
- [README.md - Troubleshooting](README.md#troubleshooting)
- [examples/DOCKER_DEPLOYMENT.md - Troubleshooting](examples/DOCKER_DEPLOYMENT.md#troubleshooting)
- [examples/CONFIGURATION.md - Troubleshooting](examples/CONFIGURATION.md#troubleshooting-configuration)
- [QUICKSTART.md - Common Issues](QUICKSTART.md#common-issues)

### Production Deployment
- [examples/DEPLOYMENT.md - Production Deployment](examples/DEPLOYMENT.md#production-deployment)
- [examples/DOCKER_DEPLOYMENT.md - Production Considerations](examples/DOCKER_DEPLOYMENT.md#production-considerations)

### Testing
- [examples/TESTING.md](examples/TESTING.md)
- [README.md - Development](README.md#development)

## Documentation Files

### Root Directory

| File | Description |
|------|-------------|
| [README.md](README.md) | Main project documentation |
| [QUICKSTART.md](QUICKSTART.md) | 5-minute quick start guide |
| [DOCUMENTATION.md](DOCUMENTATION.md) | This file - documentation index |
| [Dockerfile](Dockerfile) | Docker image definition |
| [docker-compose.yaml](docker-compose.yaml) | Docker Compose for MCP server |
| [docker-compose-jenkins.yaml](docker-compose-jenkins.yaml) | Docker Compose for test Jenkins |
| [.dockerignore](.dockerignore) | Docker build optimization |

### Examples Directory

| File | Description |
|------|-------------|
| [examples/README.md](examples/README.md) | Examples directory navigation |
| [examples/config.yaml](examples/config.yaml) | Example configuration file |
| [examples/.env.example](examples/.env.example) | Example environment variables |
| [examples/CONFIGURATION.md](examples/CONFIGURATION.md) | Configuration reference |
| [examples/DEPLOYMENT.md](examples/DEPLOYMENT.md) | Complete deployment guide |
| [examples/DOCKER_DEPLOYMENT.md](examples/DOCKER_DEPLOYMENT.md) | Docker deployment guide |
| [examples/TESTING.md](examples/TESTING.md) | Testing guide |

## Quick Reference

### Common Tasks

**Install the server:**
```bash
go install github.com/NithishNithi/go-jenkins-mcp/cmd/jenkins-mcp-server@latest
```

**Run with Docker:**
```bash
docker run -i \
  -e JENKINS_URL=https://jenkins.example.com \
  -e JENKINS_API_TOKEN=your-token \
  jenkins-mcp-server:latest
```

**Configure for Claude Desktop:**
```json
{
  "mcpServers": {
    "jenkins": {
      "command": "/usr/local/bin/jenkins-mcp-server",
      "env": {
        "JENKINS_URL": "https://jenkins.example.com",
        "JENKINS_API_TOKEN": "your-token"
      }
    }
  }
}
```

**Run local Jenkins for testing:**
```bash
docker-compose -f docker-compose-jenkins.yaml up -d
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
