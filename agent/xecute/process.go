package xecute

type LogLevel = string

const (
	LogDetailed = "detailed"
	LogNormal   = "normal"
)

type Status = string

const (
	// Process & management routine not running.
	StatusStopped = "stopped"

	// Management routine running, process not up yet.
	StatusStarting = "starting"

	// Process and management routine running and healthy.
	StatusHealthy = "healthy"

	// Process stopping, management routine still running.
	StatusStopping = "stopping"

	// Process and management routine stopped because of an error.
	StatusError = "error"
)

type ProcessConfig interface {
	HealthCheckURL() string
}
