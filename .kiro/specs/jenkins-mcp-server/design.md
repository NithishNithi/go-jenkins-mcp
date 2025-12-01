# Design Document: Jenkins MCP Server

## Overview

The Jenkins MCP Server is a Go-based implementation of the Model Context Protocol (MCP) that provides a standardized interface for AI assistants to interact with Jenkins CI/CD systems. The server acts as a bridge between MCP clients and Jenkins instances, translating MCP tool calls into Jenkins API requests and formatting responses according to the MCP specification.

The architecture follows a layered approach with clear separation between the MCP protocol layer, business logic layer, and Jenkins API client layer. This design ensures maintainability, testability, and extensibility.

## Architecture

### High-Level Architecture

```
┌─────────────────┐
│   MCP Client    │
│  (AI Assistant) │
└────────┬────────┘
         │ MCP Protocol (stdio/HTTP)
         │
┌────────▼────────────────────────────┐
│      MCP Server Layer               │
│  - Protocol handling                │
│  - Tool registration                │
│  - Request/response formatting      │
└────────┬────────────────────────────┘
         │
┌────────▼────────────────────────────┐
│      Business Logic Layer           │
│  - Input validation                 │
│  - Error handling                   │
│  - Response transformation          │
└────────┬────────────────────────────┘
         │
┌────────▼────────────────────────────┐
│      Jenkins Client Layer           │
│  - HTTP client                      │
│  - Authentication                   │
│  - API request construction         │
└────────┬────────────────────────────┘
         │ Jenkins REST API
         │
┌────────▼────────────┐
│  Jenkins Instance   │
└─────────────────────┘
```

### Component Interaction Flow

1. MCP Client sends tool invocation request via stdio
2. MCP Server Layer parses and validates the MCP request
3. Business Logic Layer validates inputs and orchestrates the operation
4. Jenkins Client Layer makes authenticated API calls to Jenkins
5. Response flows back through layers with appropriate transformations
6. MCP Server Layer formats and returns MCP-compliant response

## Components and Interfaces

### 1. MCP Server Component

**Responsibilities:**
- Implement MCP protocol handshake and communication
- Register available tools with schemas
- Parse incoming MCP requests
- Format outgoing MCP responses
- Handle MCP-level errors

**Key Interfaces:**
```go
type MCPServer interface {
    Start() error
    Stop() error
    RegisterTool(tool Tool) error
    HandleRequest(request MCPRequest) (MCPResponse, error)
}

type Tool interface {
    Name() string
    Description() string
    Schema() ToolSchema
    Execute(params map[string]interface{}) (interface{}, error)
}
```

### 2. Jenkins Client Component

**Responsibilities:**
- Manage HTTP connections to Jenkins
- Handle authentication (username/password, API token)
- Construct Jenkins API requests
- Parse Jenkins API responses
- Handle Jenkins-specific errors

**Key Interfaces:**
```go
type JenkinsClient interface {
    // Job operations
    ListJobs(folder string) ([]Job, error)
    GetJob(jobName string) (*JobDetails, error)
    
    // Build operations
    TriggerBuild(jobName string, params map[string]string) (*QueueItem, error)
    GetBuild(jobName string, buildNumber int) (*Build, error)
    GetLatestBuild(jobName string) (*Build, error)
    StopBuild(jobName string, buildNumber int) error
    
    // Log and artifact operations
    GetBuildLog(jobName string, buildNumber int) (string, error)
    ListArtifacts(jobName string, buildNumber int) ([]Artifact, error)
    GetArtifact(jobName string, buildNumber int, artifactPath string) ([]byte, error)
    
    // Queue operations
    GetQueue() ([]QueueItem, error)
}
```

### 3. Tool Implementations

Each Jenkins operation is exposed as an MCP tool:

- `jenkins_list_jobs` - List available Jenkins jobs
- `jenkins_get_job` - Get detailed job information
- `jenkins_trigger_build` - Trigger a new build
- `jenkins_get_build` - Get build status and details
- `jenkins_get_build_log` - Retrieve build console output
- `jenkins_list_artifacts` - List build artifacts
- `jenkins_get_artifact` - Download a specific artifact
- `jenkins_get_queue` - Get current build queue
- `jenkins_stop_build` - Stop a running build

### 4. Configuration Component

**Responsibilities:**
- Load configuration from environment variables or file
- Validate configuration values
- Provide configuration to other components

**Configuration Structure:**
```go
type Config struct {
    JenkinsURL      string
    Username        string
    Password        string
    APIToken        string
    Timeout         time.Duration
    TLSSkipVerify   bool
    CACertPath      string
    MaxRetries      int
    RetryBackoff    time.Duration
}
```

## Data Models

### Job Models

```go
type Job struct {
    Name        string
    URL         string
    Description string
    Buildable   bool
    InQueue     bool
    Color       string // Indicates status
}

type JobDetails struct {
    Job
    LastBuild       *BuildReference
    LastSuccessfulBuild *BuildReference
    LastFailedBuild *BuildReference
    Parameters      []JobParameter
    Disabled        bool
}

type JobParameter struct {
    Name         string
    Type         string
    DefaultValue interface{}
    Description  string
}
```

### Build Models

```go
type Build struct {
    Number      int
    URL         string
    Result      string // SUCCESS, FAILURE, ABORTED, etc.
    Building    bool
    Duration    int64
    Timestamp   int64
    Executor    string
    EstimatedDuration int64
}

type BuildReference struct {
    Number int
    URL    string
}

type QueueItem struct {
    ID          int
    JobName     string
    Why         string
    Blocked     bool
    Buildable   bool
    Stuck       bool
    InQueueSince int64
    Parameters  map[string]string
}
```

### Artifact Models

```go
type Artifact struct {
    FileName     string
    RelativePath string
    Size         int64
}
```

## C
orrectness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

After reviewing all acceptance criteria, I've identified properties that can be combined or are redundant. Many criteria about "including required fields" can be consolidated into comprehensive data completeness properties. Similarly, error handling properties can be grouped by error type rather than having separate properties for each operation.

### Authentication and Connection Properties

**Property 1: Invalid credentials produce clear error messages**
*For any* set of invalid credentials (wrong password, invalid token, malformed credentials), authentication attempts should return error messages that clearly indicate authentication failure.
**Validates: Requirements 1.2**

**Property 2: Invalid connection parameters are validated**
*For any* invalid connection parameter (malformed URL, empty required fields, invalid port), the server should validate and return specific error messages identifying the validation failure.
**Validates: Requirements 1.4**

### Job Listing and Details Properties

**Property 3: Job list completeness**
*For any* Jenkins instance state, listing jobs should return all jobs accessible to the authenticated user with complete information (name, description, status).
**Validates: Requirements 2.1, 2.2, 2.3**

**Property 4: Folder-based job listing**
*For any* folder path, listing jobs in that folder should return only jobs within that specific folder.
**Validates: Requirements 2.4**

**Property 5: Job details completeness**
*For any* valid job, retrieving job details should return complete information including last build status, build number, timestamp, and if parameterized, all parameter details with types and default values.
**Validates: Requirements 3.1, 3.3, 3.4**

**Property 6: Disabled job status indication**
*For any* disabled job, the job details response should correctly indicate the disabled status.
**Validates: Requirements 3.5**

### Build Triggering Properties

**Property 7: Build trigger returns queue identifier**
*For any* valid job and valid parameters, triggering a build should return a queue item identifier.
**Validates: Requirements 4.1, 4.3**

**Property 8: Parameter passing integrity**
*For any* parameterized job and set of valid parameters, triggering a build should pass all parameters to Jenkins without modification.
**Validates: Requirements 4.2**

**Property 9: Invalid parameter validation**
*For any* set of invalid parameters (wrong type, missing required, out of range), the server should validate and return specific error messages.
**Validates: Requirements 4.5**

### Build Status and Information Properties

**Property 10: Build information completeness**
*For any* valid job and build number, retrieving build information should include all required fields: result, duration, timestamp, executor information, and building status.
**Validates: Requirements 5.1, 5.2**

**Property 11: In-progress build status indication**
*For any* build that is currently running, the build information should indicate in-progress status and include estimated remaining time if available.
**Validates: Requirements 5.3**

### Build Log Properties

**Property 12: Build log retrieval**
*For any* valid job and build number, retrieving logs should return the complete console output available at that time.
**Validates: Requirements 6.1, 6.2**

**Property 13: Log formatting preservation**
*For any* build log content with formatting and line breaks, retrieving the log should preserve all formatting characters and line breaks.
**Validates: Requirements 6.5**

### Artifact Properties

**Property 14: Artifact list completeness**
*For any* build with artifacts, listing artifacts should return all artifacts with complete information (name, size, relative path).
**Validates: Requirements 7.1**

**Property 15: Artifact content integrity**
*For any* artifact, downloading it should return content that matches the artifact stored in Jenkins (byte-for-byte equality).
**Validates: Requirements 7.2**

### Build Queue Properties

**Property 16: Queue information completeness**
*For any* Jenkins build queue state, retrieving the queue should return all items with complete information including job name, queue position, wait time, and if parameterized, parameter information.
**Validates: Requirements 8.1, 8.2, 8.4**

**Property 17: Blocked queue item indication**
*For any* blocked queue item, the queue information should include the reason for blocking.
**Validates: Requirements 8.5**

### Build Control Properties

**Property 18: Build stop state transition**
*For any* running build, stopping it should result in the build status being updated to aborted and confirmation being returned.
**Validates: Requirements 9.1, 9.2, 9.5**

### MCP Protocol Compliance Properties

**Property 19: MCP request parsing**
*For any* valid MCP request according to the specification, the server should successfully parse and validate it.
**Validates: Requirements 10.2**

**Property 20: MCP response formatting**
*For any* operation result, the server should format the response according to the MCP specification with proper structure and fields.
**Validates: Requirements 10.3**

**Property 21: MCP error response compliance**
*For any* error condition, the server should return an MCP-compliant error response with appropriate error codes and messages.
**Validates: Requirements 10.4**

**Property 22: Tool schema validity**
*For any* exposed tool, the tool registration should include a valid schema and description conforming to MCP requirements.
**Validates: Requirements 10.5**

### Error Handling Properties

**Property 23: Retry behavior with exponential backoff**
*For any* transient network failure, the server should retry the operation with exponential backoff up to the configured maximum retries.
**Validates: Requirements 11.1**

**Property 24: Jenkins error message transformation**
*For any* Jenkins API error response, the server should parse it and return a meaningful error message to the client.
**Validates: Requirements 11.2**

**Property 25: Input validation error specificity**
*For any* invalid input across all operations, the server should validate and return specific error messages identifying what validation failed.
**Validates: Requirements 11.3**

### Configuration Properties

**Property 26: URL format validation**
*For any* configuration with a Jenkins URL, the server should validate the URL format and reject malformed URLs with specific error messages.
**Validates: Requirements 12.2**

**Property 27: Timeout application**
*For any* configured timeout value, the server should apply it to all Jenkins API requests.
**Validates: Requirements 12.4**

## Error Handling

### Error Categories

1. **Network Errors**
   - Connection failures
   - Timeouts
   - DNS resolution failures
   - TLS/SSL errors

2. **Authentication Errors**
   - Invalid credentials
   - Expired tokens
   - Insufficient permissions

3. **Validation Errors**
   - Invalid input parameters
   - Malformed requests
   - Missing required fields

4. **Jenkins API Errors**
   - Job not found
   - Build not found
   - Invalid job configuration
   - Jenkins internal errors

5. **MCP Protocol Errors**
   - Invalid MCP request format
   - Unsupported MCP version
   - Unknown tool invocation

### Error Handling Strategy

**Retry Logic:**
- Implement exponential backoff for transient failures
- Maximum 3 retry attempts by default (configurable)
- Only retry idempotent operations (GET requests)
- Do not retry authentication failures or validation errors

**Error Response Format:**
All errors returned to MCP clients follow this structure:
```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**Error Codes:**
- `AUTH_FAILED` - Authentication failure
- `NOT_FOUND` - Resource not found
- `INVALID_INPUT` - Input validation failure
- `NETWORK_ERROR` - Network connectivity issue
- `TIMEOUT` - Operation timeout
- `PERMISSION_DENIED` - Insufficient permissions
- `JENKINS_ERROR` - Jenkins API error
- `INTERNAL_ERROR` - Unexpected server error

## Testing Strategy

### Unit Testing

Unit tests will verify specific examples and integration points:

- **Configuration loading**: Test loading from environment variables and files
- **Authentication methods**: Test both username/password and API token auth
- **MCP protocol handshake**: Test successful handshake sequence
- **TLS/SSL connections**: Test secure connections with valid certificates
- **Latest build query**: Test querying without specifying build number
- **Empty result handling**: Test empty job lists, empty queues, no artifacts
- **Error edge cases**: Test non-existent jobs, non-existent builds, stopped builds, authorization failures, missing logs, missing artifacts, timeout scenarios

### Property-Based Testing

We will use the `gopter` library for property-based testing in Go. Each property test will run a minimum of 100 iterations to ensure thorough coverage.

**Property Test Configuration:**
```go
parameters := gopter.DefaultTestParameters()
parameters.MinSuccessfulTests = 100
```

**Test Generators:**
We will implement custom generators for:
- Valid and invalid credentials
- Valid and invalid URLs
- Job configurations with various parameter types
- Build states (running, completed, failed, aborted)
- Queue items with different states
- Log content with various formatting
- Artifact metadata

**Property Test Organization:**
Each correctness property from the design document will be implemented as a single property-based test, tagged with a comment in this format:
```go
// Feature: jenkins-mcp-server, Property 1: Invalid credentials produce clear error messages
func TestProperty1_InvalidCredentialsError(t *testing.T) {
    // Property test implementation
}
```

**Key Property Tests:**
- Authentication and validation properties (Properties 1-2)
- Data completeness properties (Properties 3-6, 10-11, 14, 16)
- Data integrity properties (Properties 8, 13, 15)
- State transition properties (Property 18)
- Protocol compliance properties (Properties 19-22)
- Error handling properties (Properties 23-25)
- Configuration properties (Properties 26-27)

### Integration Testing

Integration tests will verify end-to-end flows with a real Jenkins instance (using Docker for test environment):

- Complete workflow: list jobs → get job details → trigger build → monitor status → retrieve logs
- Error recovery: network interruption → retry → success
- Permission scenarios: different user roles and access levels

### Test Utilities

We will create test utilities for:
- Mock Jenkins API server for unit tests
- Test data generators for property tests
- Docker-based Jenkins instance for integration tests
- MCP client simulator for protocol testing

## Implementation Notes

### Go Libraries

- **MCP SDK**: Use `github.com/modelcontextprotocol/go-sdk` for MCP protocol implementation
- **HTTP Client**: Use standard `net/http` with custom transport for retry logic
- **JSON Parsing**: Use standard `encoding/json`
- **Configuration**: Use `github.com/spf13/viper` for flexible configuration management
- **Logging**: Use `github.com/sirupsen/logrus` for structured logging
- **Property Testing**: Use `github.com/leanovate/gopter` for property-based tests

### Security Considerations

- Store credentials securely (never log passwords or tokens)
- Support TLS certificate verification
- Implement request timeout to prevent hanging
- Validate all user inputs before processing
- Use secure defaults (TLS verification enabled by default)

### Performance Considerations

- Implement connection pooling for HTTP client
- Stream large artifacts instead of loading into memory
- Implement pagination for large job lists
- Cache job metadata with TTL to reduce API calls
- Use context for cancellation and timeout propagation

### Extensibility

The design supports future extensions:
- Additional Jenkins API endpoints can be added as new tools
- Support for Jenkins plugins (Blue Ocean, Pipeline, etc.)
- Webhook support for build notifications
- Multi-Jenkins instance support
- Custom authentication mechanisms

## Deployment

### Configuration

The server can be configured via:

1. **Environment Variables:**
```bash
JENKINS_URL=https://jenkins.example.com
JENKINS_USERNAME=admin
JENKINS_API_TOKEN=abc123...
JENKINS_TIMEOUT=30s
JENKINS_TLS_SKIP_VERIFY=false
JENKINS_CA_CERT=/path/to/ca.crt
```

2. **Configuration File (config.yaml):**
```yaml
jenkins:
  url: https://jenkins.example.com
  username: admin
  apiToken: abc123...
  timeout: 30s
  tls:
    skipVerify: false
    caCert: /path/to/ca.crt
  retry:
    maxAttempts: 3
    backoff: 1s
```

### Running the Server

```bash
# Using environment variables
export JENKINS_URL=https://jenkins.example.com
export JENKINS_API_TOKEN=your-token
./jenkins-mcp-server

# Using configuration file
./jenkins-mcp-server --config config.yaml

# With MCP client (e.g., Claude Desktop)
# Add to MCP client configuration:
{
  "mcpServers": {
    "jenkins": {
      "command": "/path/to/jenkins-mcp-server",
      "env": {
        "JENKINS_URL": "https://jenkins.example.com",
        "JENKINS_API_TOKEN": "your-token"
      }
    }
  }
}
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o jenkins-mcp-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/jenkins-mcp-server /usr/local/bin/
ENTRYPOINT ["jenkins-mcp-server"]
```

## Future Enhancements

1. **Advanced Pipeline Support**: Support for Blue Ocean API and advanced pipeline operations
2. **Webhook Integration**: Receive build notifications via webhooks
3. **Multi-Instance Support**: Manage multiple Jenkins instances from a single MCP server
4. **Build Comparison**: Compare builds and show diffs
5. **Metrics and Monitoring**: Expose Prometheus metrics for server monitoring
6. **Caching Layer**: Implement intelligent caching to reduce Jenkins API load
7. **Plugin Support**: Extensible plugin system for custom Jenkins integrations
