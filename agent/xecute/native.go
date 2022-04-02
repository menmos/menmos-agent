package xecute

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"go.uber.org/zap"
)

/// Manages a native xecute process.
type Native struct {
	// Executable information.
	binaryPath string
	workdir    string
	cmd        *exec.Cmd
	logWriter  *logWriter
	port       uint16

	// Management stuff
	logger *zap.SugaredLogger
	stop   chan bool
	status Status
}

func NewNativeProcess(workdir, binaryPath string, logger *zap.Logger) (*Native, error) {
	logPath := path.Join(workdir, "log.json")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	logWriter := newLogWriter(logFile)

	// Allocate a port for our process.
	port, err := getFreePort()
	if err != nil {
		return nil, err
	}

	return &Native{
		binaryPath: binaryPath,
		workdir:    workdir,
		cmd:        nil,
		logWriter:  logWriter,
		port:       port,

		logger: logger.Sugar(),
		stop:   make(chan bool),
		status: StatusStopped,
	}, nil
}

func (p *Native) setStatus(status Status) {
	p.status = status
	p.logger.Infof("setting status to '%v'", status)
}

func (p *Native) stateWatcher(logLevel LogLevel, configPath string) {
	p.setStatus(StatusStarting)

	defer func() {
		p.stop <- true
	}()

	// Build the command.
	p.cmd = exec.Command(p.binaryPath, "--cfg", configPath)

	// Redirect both outputs to the log file.
	p.cmd.Stdout = p.logWriter
	p.cmd.Stderr = p.logWriter

	// Set the log level to the requested level.
	p.cmd.Env = append(p.cmd.Env, fmt.Sprintf("MENMOS_LOG_LEVEL=%s", logLevel))
	p.cmd.Env = append(p.cmd.Env, "MENMOS_LOG_JSON=true")
	p.cmd.Env = append(p.cmd.Env, fmt.Sprintf("MENMOS_SERVER_PORT=%d", p.port))

	// Start the process
	p.logger.Debugf("starting the process")
	if err := p.cmd.Start(); err != nil {
		p.logger.Errorf("failed to start process: %v", err)
		p.setStatus(StatusError)
		return
	}

	// All xecute processes have the same health URL.
	// FIXME(prod): support HTTPS if required
	healthCheckURL := fmt.Sprintf("http://localhost:%d/health", p.port)

	retry := 100
	for {
		p.logger.Debug("checking if process is healthy")
		resp, err := http.Get(healthCheckURL)
		if err != nil || resp.StatusCode != 200 {
			p.logger.Debug("process is not up yet")
			retry -= 1
			if retry == 0 {
				p.setStatus(StatusError)
				p.logger.Error("retries exceeded: process failed to come up")
				return
			}

			time.Sleep(100 * time.Millisecond)
			continue
		}

		p.setStatus(StatusHealthy)
		break
	}

	// We wait for the process to stop - either from a crash or from a stop signal.
	if err := p.cmd.Wait(); err != nil {
		p.setStatus(StatusError)
		return
	}

	if p.cmd.ProcessState.ExitCode() != 0 {
		p.setStatus(StatusError)
	} else {
		p.setStatus(StatusStopped)
	}
}

func (p *Native) Start(logLevel LogLevel) error {
	configPath := path.Join(p.workdir, "config.toml")

	go p.stateWatcher(logLevel, configPath)

	return nil
}

func (p *Native) Stop() error {
	if p.status != StatusHealthy && p.status == StatusStarting {
		return nil // We're already stopped
	}

	if p.cmd == nil || p.cmd.Process == nil {
		p.logger.Info("process never started. maybe a crash?")
		return nil
	}

	p.logger.Info("asking nicely for process to quit")
	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		return err
	}

	timer := time.AfterFunc(10*time.Second, func() {
		p.logger.Info("asking rudely for process to quit")
		p.cmd.Process.Kill()
	})
	<-p.stop
	timer.Stop()

	return nil
}

func (p *Native) GetLogs(numberOfLines uint) []interface{} {
	return p.logWriter.GetLastNLines(int(numberOfLines))
}

func (p *Native) Status() string {
	return p.status
}

func (p *Native) Port() uint16 {
	return p.port
}
