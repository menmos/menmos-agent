package process

type ProcessConfig interface {
	HealthCheckURL() string
}
