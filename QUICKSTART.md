# Quick Start Guide

Get the Jenkins MCP Server up and running in 5 minutes.

## Prerequisites

- Access to a Jenkins instance
- Jenkins API token (recommended) or username/password

## Step 1: Get Jenkins API Token

1. Log in to your Jenkins instance
2. Click your username (top right) → **Configure**
3. Scroll to **API Token** section
4. Click **Add new Token**
5. Give it a name and click **Generate**
6. **Copy the token** (you won't see it again!)

## Step 2: Choose Your Installation Method

### Option A: Docker (Recommended)

```bash
# Run with environment variables
docker run -i \
  -e JENKINS_URL=https://your-jenkins.com \
  -e JENKINS_API_TOKEN=your-token-here \
  ghcr.io/nithishnithi/jenkins-mcp-server:latest
```

### Option B: Docker Compose

```bash
# Clone the repository
git clone https://github.com/NithishNithi/go-jenkins-mcp.git
cd go-jenkins-mcp

# Create .env file
cat > .env << EOF
JENKINS_URL=https://your-jenkins.com
JENKINS_API_TOKEN=your-token-here
EOF

# Start the server
docker-compose up -d

# View logs
docker-compose logs -f
```

### Option C: Local Binary

```bash
# Install from source
go install github.com/NithishNithi/go-jenkins-mcp/cmd/jenkins-mcp-server@latest

# Set environment variables
export JENKINS_URL=https://your-jenkins.com
export JENKINS_API_TOKEN=your-token-here

# Run the server
jenkins-mcp-server
```

## Step 3: Integrate with Claude Desktop

1. **Find your Claude config file:**
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. **Add the Jenkins MCP Server:**

```json
{
  "mcpServers": {
    "jenkins": {
      "command": "/usr/local/bin/jenkins-mcp-server",
      "env": {
        "JENKINS_URL": "https://your-jenkins.com",
        "JENKINS_API_TOKEN": "your-token-here"
      }
    }
  }
}
```

Or with Docker:

```json
{
  "mcpServers": {
    "jenkins": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "JENKINS_URL=https://your-jenkins.com",
        "-e", "JENKINS_API_TOKEN=your-token-here",
        "jenkins-mcp-server:latest"
      ]
    }
  }
}
```

3. **Restart Claude Desktop**

## Step 4: Test It Out

Ask Claude:
- "List all Jenkins jobs"
- "What's the status of the latest build for [job-name]?"
- "Trigger a build for [job-name]"
- "Show me the build log for build #42 of [job-name]"

## Common Issues

### "Connection refused"
- Check that `JENKINS_URL` is correct and accessible
- Verify Jenkins is running

### "Authentication failed"
- Verify your API token is correct
- Try regenerating the token
- Check user permissions in Jenkins

### "Certificate verification failed"
For development with self-signed certificates:
```bash
export JENKINS_TLS_SKIP_VERIFY=true
```
**Warning:** Don't use this in production!

### Claude doesn't show Jenkins tools
- Verify the binary path is correct
- Check environment variables are set
- Restart Claude Desktop
- Check Claude Desktop logs

## Next Steps

- **Full Documentation**: See [README.md](README.md)
- **Configuration Options**: See [examples/CONFIGURATION.md](examples/CONFIGURATION.md)
- **Docker Deployment**: See [examples/DOCKER_DEPLOYMENT.md](examples/DOCKER_DEPLOYMENT.md)
- **Production Setup**: See [examples/DEPLOYMENT.md](examples/DEPLOYMENT.md)

## Available Tools

Once configured, you can use these Jenkins operations through Claude:

- `jenkins_list_jobs` - List all jobs
- `jenkins_get_job` - Get job details
- `jenkins_trigger_build` - Trigger a build
- `jenkins_get_build` - Get build status
- `jenkins_get_build_log` - Get build logs
- `jenkins_list_artifacts` - List build artifacts
- `jenkins_get_artifact` - Download artifacts
- `jenkins_get_queue` - View build queue
- `jenkins_stop_build` - Stop a running build

## Support

- **Issues**: [GitHub Issues](https://github.com/NithishNithi/go-jenkins-mcp/issues)
- **Documentation**: [Full README](README.md)
- **Examples**: [examples/](examples/)

---

**That's it!** You're now ready to use Jenkins through Claude or any other MCP client.
