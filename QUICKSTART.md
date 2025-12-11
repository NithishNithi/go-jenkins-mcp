# Quick Start Guide

Get the Jenkins MCP Server up and running in 5 minutes.

## Prerequisites

- Access to a Jenkins instance
- Jenkins API token

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
docker run -i \
  -e JENKINS_URL=https://your-jenkins.com \
  -e JENKINS_USERNAME=your-username \
  -e JENKINS_API_TOKEN=your-token-here \
  -e JENKINS_TOOL_PREFIX=prod \
  ghcr.io/nithishnithi/jenkins-mcp-server:latest
```

**Note:** `JENKINS_TOOL_PREFIX` is optional but useful when running multiple Jenkins MCP servers (e.g., `prod_jenkins_list_jobs`, `staging_jenkins_list_jobs`).

### Option B: Docker Compose

```bash
# Clone the repository
git clone https://github.com/NithishNithi/go-jenkins-mcp.git
cd go-jenkins-mcp

# Create .env file
cat > .env << EOF
JENKINS_URL=https://your-jenkins.com
JENKINS_USERNAME=your-username
JENKINS_API_TOKEN=your-token-here
JENKINS_TOOL_PREFIX=prod
EOF

# Start the server
docker-compose up -d

# View logs
docker-compose logs -f
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
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "JENKINS_URL=https://your-jenkins.com",
        "-e", "JENKINS_USERNAME=your-username",
        "-e", "JENKINS_API_TOKEN=your-api-token",
        "-e", "JENKINS_TOOL_PREFIX=prod",
        "ghcr.io/nithishnithi/jenkins-mcp-server:latest"
      ]
    }
  }
}
```

3. **Restart Claude Desktop**

## Step 4: Test It Out

Ask Claude:
- "List all Jenkins jobs"
- "What's the status of the latest build for main-pipeline?"
- "Trigger a build for deployment-job"
- "Show me the build log for build #42"

## Troubleshooting

### "Connection refused"
- Verify `JENKINS_URL` is correct and accessible
- Check that Jenkins is running

### "Authentication failed"
- Verify your API token is correct
- Try regenerating the token
- Check user permissions in Jenkins

### "Certificate verification failed"
For self-signed certificates (development only):
```bash
-e JENKINS_TLS_SKIP_VERIFY=true
```
**Warning:** Don't use in production!

### Claude doesn't show Jenkins tools
- Verify configuration is correct
- Check environment variables are set
- Restart Claude Desktop
- Ensure binary has execute permissions

## Available Tools

**Jobs:** list_jobs, get_job, trigger_build  
**Builds:** get_build, get_build_log, get_running_builds, stop_build  
**Artifacts:** list_artifacts, get_artifact  
**Queue:** get_queue, get_queue_item, cancel_queue_item  
**Views:** list_views, get_view, create_view  
**Server:** server_health, list_nodes, get_pipeline_script

## Next Steps

- **Full Documentation**: [README.md](README.md)
- **Issues**: [GitHub Issues](https://github.com/NithishNithi/go-jenkins-mcp/issues)

---

**That's it!** You're now ready to use Jenkins through Claude.