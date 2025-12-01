package jenkins

// Job represents basic job information
type Job struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Buildable   bool   `json:"buildable"`
	InQueue     bool   `json:"inQueue"`
	Color       string `json:"color"` // Indicates status
}

// JobDetails represents detailed job information
type JobDetails struct {
	Job
	LastBuild           *BuildReference `json:"lastBuild,omitempty"`
	LastSuccessfulBuild *BuildReference `json:"lastSuccessfulBuild,omitempty"`
	LastFailedBuild     *BuildReference `json:"lastFailedBuild,omitempty"`
	Parameters          []JobParameter  `json:"parameters,omitempty"`
	Disabled            bool            `json:"disabled"`
}

// JobParameter represents a job parameter definition
type JobParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Description  string      `json:"description,omitempty"`
}

// Build represents build information
type Build struct {
	Number            int    `json:"number"`
	URL               string `json:"url"`
	Result            string `json:"result"` // SUCCESS, FAILURE, ABORTED, etc.
	Building          bool   `json:"building"`
	Duration          int64  `json:"duration"`
	Timestamp         int64  `json:"timestamp"`
	Executor          string `json:"executor,omitempty"`
	EstimatedDuration int64  `json:"estimatedDuration,omitempty"`
}

// BuildReference represents a reference to a build
type BuildReference struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

// QueueItem represents an item in the build queue
type QueueItem struct {
	ID           int               `json:"id"`
	JobName      string            `json:"jobName"`
	Why          string            `json:"why"`
	Blocked      bool              `json:"blocked"`
	Buildable    bool              `json:"buildable"`
	Stuck        bool              `json:"stuck"`
	InQueueSince int64             `json:"inQueueSince"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

// Artifact represents a build artifact
type Artifact struct {
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
	Size         int64  `json:"size"`
}

// RunningBuild represents a currently running build
type RunningBuild struct {
	JobName           string `json:"jobName"`
	BuildNumber       int    `json:"buildNumber"`
	URL               string `json:"url"`
	Timestamp         int64  `json:"timestamp"`
	EstimatedDuration int64  `json:"estimatedDuration"`
	Executor          string `json:"executor,omitempty"`
}

// View represents a Jenkins view
type View struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// ViewDetails represents detailed view information
type ViewDetails struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	Jobs        []Job  `json:"jobs"`
}
