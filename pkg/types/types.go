package types

import (
	"os/exec"
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusRunning JobStatus = "Running"
	JobStatusStopped JobStatus = "Stopped"
	JobStatusDone    JobStatus = "Done"
)

// Job represents a background job
type Job struct {
	ID         int
	PID        int
	Command    string
	Args       []string
	Status     JobStatus
	Cmd        *exec.Cmd
	StartTime  time.Time
	EndTime    *time.Time
	Background bool
}
