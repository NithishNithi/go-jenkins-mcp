# Implementation Plan

- [x] 1. Set up project structure and dependencies
  - Initialize Go module with appropriate name (e.g., github.com/yourusername/jenkins-mcp-server)
  - Add required dependencies: MCP SDK (github.com/modelcontextprotocol/go-sdk), viper, logrus, gopter
  - Create directory structure: cmd/jenkins-mcp-server/, internal/jenkins/, internal/mcp/, internal/config/
  - Set up basic main.go entry point with stdio communication
  - _Requirements: 10.1, 12.1_

- [x] 2. Implement configuration management
  - [x] 2.1 Create configuration data structures
    - Define Config struct with fields: JenkinsURL, Username, Password, APIToken, Timeout, TLSSkipVerify, CACertPath, MaxRetries, RetryBackoff
    - Implement validation methods for configuration values (URL format, required fields, timeout ranges)
    - _Requirements: 12.1, 12.2_
  
  - [x] 2.2 Implement configuration loading
    - Write code to load from environment variables (JENKINS_URL, JENKINS_USERNAME, JENKINS_PASSWORD, JENKINS_API_TOKEN, etc.)
    - Write code to load from configuration file using viper
    - Implement configuration merging logic (file overrides defaults, env vars override file)
    - _Requirements: 12.1_
  
  - [ ]* 2.3 Write property test for URL validation
    - **Property 26: URL format validation**
    - **Validates: Requirements 12.2**
  
  - [ ]* 2.4 Write unit tests for configuration
    - Test loading from environment variables
    - Test loading from configuration file
    - Test both authentication methods (username/password and API token)
    - _Requirements: 12.1, 12.3_

- [x] 3. Implement Jenkins client layer
  - [x] 3.1 Create Jenkins client interface and HTTP client setup
    - Define JenkinsClient interface with all methods (ListJobs, GetJob, TriggerBuild, GetBuild, GetLatestBuild, StopBuild, GetBuildLog, ListArtifacts, GetArtifact, GetQueue)
    - Implement HTTP client with connection pooling and custom transport
    - Implement authentication (basic auth for username/password, bearer token for API token)
    - Add timeout configuration and context propagation
    - Implement retry logic with exponential backoff for transient failures
    - _Requirements: 1.1, 1.5, 12.4_
  
  - [ ]* 3.2 Write property test for invalid credentials
    - **Property 1: Invalid credentials produce clear error messages**
    - **Validates: Requirements 1.2**
  
  - [ ]* 3.3 Write property test for invalid connection parameters
    - **Property 2: Invalid connection parameters are validated**
    - **Validates: Requirements 1.4**
  
  - [ ]* 3.4 Write unit tests for authentication and connection
    - Test successful connection with valid credentials
    - Test TLS/SSL connections
    - Test connection to unreachable Jenkins instance
    - _Requirements: 1.1, 1.3, 1.5_

- [x] 4. Implement job operations
  - [x] 4.1 Implement job listing functionality
    - Write ListJobs method with optional folder parameter
    - Make GET request to /api/json endpoint with tree parameter for job data
    - Parse Jenkins API response for job list (name, url, description, buildable, inQueue, color)
    - Handle permission filtering (only return accessible jobs)
    - _Requirements: 2.1, 2.3, 2.4_
  
  - [x] 4.2 Implement job details retrieval
    - Write GetJob method that takes job name
    - Make GET request to /job/{jobName}/api/json with detailed tree parameter
    - Parse complete job information including lastBuild, lastSuccessfulBuild, lastFailedBuild, parameters, disabled status
    - Handle job parameter types (string, boolean, choice, etc.)
    - _Requirements: 3.1, 3.3, 3.4, 3.5_
  
  - [ ]* 4.3 Write property test for job list completeness
    - **Property 3: Job list completeness**
    - **Validates: Requirements 2.1, 2.2, 2.3**
  
  - [ ]* 4.4 Write property test for folder-based job listing
    - **Property 4: Folder-based job listing**
    - **Validates: Requirements 2.4**
  
  - [ ]* 4.5 Write property test for job details completeness
    - **Property 5: Job details completeness**
    - **Validates: Requirements 3.1, 3.3, 3.4**
  
  - [ ]* 4.6 Write property test for disabled job status
    - **Property 6: Disabled job status indication**
    - **Validates: Requirements 3.5**
  
  - [ ]* 4.7 Write unit tests for job operations
    - Test empty job list handling
    - Test non-existent job error handling
    - _Requirements: 2.5, 3.2_

- [x] 5. Implement build triggering operations
  - [x] 5.1 Implement build trigger functionality
    - Write TriggerBuild method that takes job name and parameters map
    - Make POST request to /job/{jobName}/build or /job/{jobName}/buildWithParameters
    - Handle parameterized builds by encoding parameters in request body
    - Parse queue item response from Location header
    - Validate parameters against job definition before triggering
    - _Requirements: 4.1, 4.2, 4.3_
  
  - [ ]* 5.2 Write property test for build trigger
    - **Property 7: Build trigger returns queue identifier**
    - **Validates: Requirements 4.1, 4.3**
  
  - [ ]* 5.3 Write property test for parameter passing
    - **Property 8: Parameter passing integrity**
    - **Validates: Requirements 4.2**
  
  - [ ]* 5.4 Write property test for invalid parameter validation
    - **Property 9: Invalid parameter validation**
    - **Validates: Requirements 4.5**
  
  - [ ]* 5.5 Write unit tests for build triggering
    - Test authorization error handling
    - _Requirements: 4.4_

- [x] 6. Implement build status and information operations
  - [x] 6.1 Implement build information retrieval
    - Write GetBuild method that takes job name and build number
    - Make GET request to /job/{jobName}/{buildNumber}/api/json
    - Write GetLatestBuild method that queries lastBuild
    - Parse build information: number, url, result, building, duration, timestamp, executor, estimatedDuration
    - Handle in-progress builds (building=true, result=null)
    - _Requirements: 5.1, 5.2, 5.3, 5.5_
  
  - [ ]* 6.2 Write property test for build information completeness
    - **Property 10: Build information completeness**
    - **Validates: Requirements 5.1, 5.2**
  
  - [ ]* 6.3 Write property test for in-progress build status
    - **Property 11: In-progress build status indication**
    - **Validates: Requirements 5.3**
  
  - [ ]* 6.4 Write unit tests for build operations
    - Test non-existent build error handling
    - Test latest build query without build number
    - _Requirements: 5.4, 5.5_

- [x] 7. Implement build log operations
  - [x] 7.1 Implement log retrieval functionality
    - Write GetBuildLog method that takes job name and build number
    - Make GET request to /job/{jobName}/{buildNumber}/consoleText
    - Handle in-progress build logs (partial content)
    - Preserve log formatting (line breaks, ANSI codes)
    - Support optional size limits for large logs
    - _Requirements: 6.1, 6.2, 6.5_
  
  - [ ]* 7.2 Write property test for build log retrieval
    - **Property 12: Build log retrieval**
    - **Validates: Requirements 6.1, 6.2**
  
  - [ ]* 7.3 Write property test for log formatting preservation
    - **Property 13: Log formatting preservation**
    - **Validates: Requirements 6.5**
  
  - [ ]* 7.4 Write unit tests for log operations
    - Test large log handling with size limits
    - Test non-existent log error handling
    - _Requirements: 6.3, 6.4_

- [x] 8. Implement artifact operations
  - [x] 8.1 Implement artifact listing and retrieval
    - Write ListArtifacts method that takes job name and build number
    - Make GET request to /job/{jobName}/{buildNumber}/api/json with artifacts tree
    - Parse artifact list: fileName, relativePath, size
    - Write GetArtifact method with streaming support
    - Make GET request to /job/{jobName}/{buildNumber}/artifact/{relativePath}
    - Handle large artifacts efficiently using io.Reader/Writer
    - _Requirements: 7.1, 7.2, 7.5_
  
  - [ ]* 8.2 Write property test for artifact list completeness
    - **Property 14: Artifact list completeness**
    - **Validates: Requirements 7.1**
  
  - [ ]* 8.3 Write property test for artifact content integrity
    - **Property 15: Artifact content integrity**
    - **Validates: Requirements 7.2**
  
  - [ ]* 8.4 Write unit tests for artifact operations
    - Test empty artifact list handling
    - Test non-existent artifact error handling
    - _Requirements: 7.3, 7.4_

- [x] 9. Implement build queue operations
  - [x] 9.1 Implement queue retrieval functionality
    - Write GetQueue method
    - Make GET request to /queue/api/json
    - Parse queue items: id, task.name (job name), why, blocked, buildable, stuck, inQueueSince, params
    - Handle blocked items with reasons from 'why' field
    - _Requirements: 8.1, 8.2, 8.4, 8.5_
  
  - [ ]* 9.2 Write property test for queue information completeness
    - **Property 16: Queue information completeness**
    - **Validates: Requirements 8.1, 8.2, 8.4**
  
  - [ ]* 9.3 Write property test for blocked queue item indication
    - **Property 17: Blocked queue item indication**
    - **Validates: Requirements 8.5**
  
  - [ ]* 9.4 Write unit tests for queue operations
    - Test empty queue handling
    - _Requirements: 8.3_

- [x] 10. Implement build control operations
  - [x] 10.1 Implement build stop functionality
    - Write StopBuild method that takes job name and build number
    - Make POST request to /job/{jobName}/{buildNumber}/stop
    - Verify status update to aborted by querying build status after stop
    - Return confirmation response with updated build status
    - _Requirements: 9.1, 9.2, 9.5_
  
  - [ ]* 10.2 Write property test for build stop state transition
    - **Property 18: Build stop state transition**
    - **Validates: Requirements 9.1, 9.2, 9.5**
  
  - [ ]* 10.3 Write unit tests for build control
    - Test stopping completed build error handling
    - Test authorization error handling
    - _Requirements: 9.3, 9.4_

- [x] 11. Implement error handling and retry logic
  - [x] 11.1 Create error types and error response structures
    - Define error codes: AUTH_FAILED, NOT_FOUND, INVALID_INPUT, NETWORK_ERROR, TIMEOUT, PERMISSION_DENIED, JENKINS_ERROR, INTERNAL_ERROR
    - Define ErrorResponse struct with Code, Message, Details fields
    - Implement error wrapping and context using Go error wrapping
    - Create helper functions for common error scenarios
    - _Requirements: 11.2_
  
  - [x] 11.2 Implement retry logic with exponential backoff
    - Create retry wrapper for HTTP requests using custom RoundTripper
    - Implement exponential backoff algorithm (initial delay, max delay, multiplier)
    - Configure retry for idempotent operations only (GET requests)
    - Add max retry attempts configuration (default 3)
    - _Requirements: 11.1_
  
  - [ ]* 11.3 Write property test for retry behavior
    - **Property 23: Retry behavior with exponential backoff**
    - **Validates: Requirements 11.1**
  
  - [ ]* 11.4 Write property test for Jenkins error transformation
    - **Property 24: Jenkins error message transformation**
    - **Validates: Requirements 11.2**
  
  - [ ]* 11.5 Write property test for input validation errors
    - **Property 25: Input validation error specificity**
    - **Validates: Requirements 11.3**
  
  - [ ]* 11.6 Write unit tests for error handling
    - Test timeout error handling
    - _Requirements: 11.4_

- [x] 12. Implement MCP server layer
  - [x] 12.1 Set up MCP server and protocol handling
    - Initialize MCP server using github.com/modelcontextprotocol/go-sdk
    - Implement protocol handshake (initialize, initialized messages)
    - Set up stdio communication (read from stdin, write to stdout)
    - Implement request routing to appropriate tool handlers
    - _Requirements: 10.1_
  
  - [x] 12.2 Implement tool registration
    - Define tool schemas for all Jenkins operations with JSON schema for parameters
    - Create schemas for: jenkins_list_jobs, jenkins_get_job, jenkins_trigger_build, jenkins_get_build, jenkins_get_build_log, jenkins_list_artifacts, jenkins_get_artifact, jenkins_get_queue, jenkins_stop_build
    - Register all tools with MCP server on startup
    - Include descriptions and parameter documentation for each tool
    - _Requirements: 10.5_
  
  - [ ]* 12.3 Write property test for MCP request parsing
    - **Property 19: MCP request parsing**
    - **Validates: Requirements 10.2**
  
  - [ ]* 12.4 Write property test for MCP response formatting
    - **Property 20: MCP response formatting**
    - **Validates: Requirements 10.3**
  
  - [ ]* 12.5 Write property test for MCP error response compliance
    - **Property 21: MCP error response compliance**
    - **Validates: Requirements 10.4**
  
  - [ ]* 12.6 Write property test for tool schema validity
    - **Property 22: Tool schema validity**
    - **Validates: Requirements 10.5**
  
  - [ ]* 12.7 Write unit tests for MCP protocol
    - Test MCP protocol handshake
    - _Requirements: 10.1_

- [x] 13. Implement MCP tool handlers
  - [x] 13.1 Implement jenkins_list_jobs tool
    - Create tool handler that extracts folder parameter from MCP request
    - Call Jenkins client ListJobs method
    - Transform response to MCP format (JSON array of job objects)
    - Handle errors and return MCP error responses with appropriate error codes
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  
  - [x] 13.2 Implement jenkins_get_job tool
    - Create tool handler that extracts jobName parameter from MCP request
    - Call Jenkins client GetJob method
    - Transform job details to MCP format (JSON object with all job fields)
    - Handle not found errors with clear messages
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  
  - [x] 13.3 Implement jenkins_trigger_build tool
    - Create tool handler that extracts jobName and parameters from MCP request
    - Parse and validate parameters against job definition
    - Call Jenkins client TriggerBuild method
    - Return queue item in MCP format (JSON object with queue ID)
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  
  - [x] 13.4 Implement jenkins_get_build tool
    - Create tool handler that extracts jobName and optional buildNumber from MCP request
    - Support both specific build (with buildNumber) and latest build (without buildNumber)
    - Call appropriate Jenkins client method (GetBuild or GetLatestBuild)
    - Transform build info to MCP format
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [x] 13.5 Implement jenkins_get_build_log tool
    - Create tool handler that extracts jobName, buildNumber, and optional sizeLimit from MCP request
    - Call Jenkins client GetBuildLog method
    - Handle log streaming and size limits
    - Return log content in MCP format (text content)
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_
  
  - [x] 13.6 Implement jenkins_list_artifacts and jenkins_get_artifact tools
    - Create jenkins_list_artifacts handler that extracts jobName and buildNumber
    - Call Jenkins client ListArtifacts method
    - Create jenkins_get_artifact handler that extracts jobName, buildNumber, and artifactPath
    - Call Jenkins client GetArtifact method with streaming support
    - Handle artifact streaming efficiently (base64 encode for MCP transport)
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  
  - [x] 13.7 Implement jenkins_get_queue tool
    - Create tool handler for queue retrieval
    - Call Jenkins client GetQueue method
    - Transform queue items to MCP format (JSON array of queue objects)
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_
  
  - [x] 13.8 Implement jenkins_stop_build tool
    - Create tool handler that extracts jobName and buildNumber from MCP request
    - Call Jenkins client StopBuild method
    - Verify state transition by querying build status
    - Return confirmation in MCP format
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ]* 14. Write property test for timeout application
  - **Property 27: Timeout application**
  - **Validates: Requirements 12.4**

- [x] 15. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 16. Create documentation and examples
  - [x] 16.1 Write README with setup instructions
    - Document installation steps (go install or binary download)
    - Provide configuration examples (environment variables and config file)
    - Include usage examples with MCP clients (Claude Desktop, other MCP clients)
    - Document all available tools and their parameters
    - Include troubleshooting section
    - _Requirements: 12.1, 12.2, 12.3, 12.5_
  
  - [x] 16.2 Create example configuration files
    - Create example config.yaml with all options documented
    - Create example .env file with environment variable examples
    - Document all configuration options: URL, auth methods, timeouts, TLS settings, retry settings
    - Include examples for both username/password and API token authentication
    - _Requirements: 12.1_
  
  - [x] 16.3 Write Dockerfile and deployment guide
    - Create Dockerfile for containerized deployment (multi-stage build)
    - Document Docker deployment steps
    - Include docker-compose.yaml example
    - Document how to pass configuration to container
    - _Requirements: 12.1_

- [ ] 17. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
