package xecute

type Status = string

const (
	// Process & management routine not running.
	StatusStopped Status = "stopped"

	// Management routine running, process not up yet.
	StatusStarting Status = "starting"

	// Process and management routine running and healthy.
	StatusHealthy Status = "healthy"

	// Process stopping, management routine still running.
	StatusStopping Status = "stopping"

	// Process and management routine stopped because of an error.
	StatusError Status = "error"
)
