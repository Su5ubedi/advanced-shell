package shell

import (
	"fmt"
	"syscall"
	"time"

	"github.com/Su5ubedi/advanced-shell/pkg/types"
)

// JobManager handles job control operations
type JobManager struct {
	jobs       map[int]*types.Job
	jobCounter int
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs:       make(map[int]*types.Job),
		jobCounter: 0,
	}
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(jobID int) (*types.Job, error) {
	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %d not found", jobID)
	}
	return job, nil
}

// GetAllJobs returns all jobs
func (jm *JobManager) GetAllJobs() []*types.Job {
	jobs := make([]*types.Job, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// ListJobs lists all jobs with their status
func (jm *JobManager) ListJobs() {
	if len(jm.jobs) == 0 {
		fmt.Println("No active jobs")
		return
	}

	fmt.Println("Active jobs:")
	for _, job := range jm.jobs {
		duration := time.Since(job.StartTime)
		if job.EndTime != nil {
			duration = job.EndTime.Sub(job.StartTime)
		}

		fmt.Printf("[%d] %s %s (PID: %d, Duration: %v)\n",
			job.ID, job.Status, job.Command, job.PID, duration.Round(time.Second))
	}
}

// BringToForeground brings a background job to the foreground
func (jm *JobManager) BringToForeground(jobID int) error {
	job, err := jm.GetJob(jobID)
	if err != nil {
		return err
	}

	if job.Status == types.JobStatusDone {
		return fmt.Errorf("job %d has already completed", jobID)
	}

	fmt.Printf("Bringing job [%d] to foreground: %s\n", job.ID, job.Command)

	// Send SIGCONT to resume the process if it's stopped
	if job.Cmd != nil && job.Cmd.Process != nil {
		if job.Status == types.JobStatusStopped {
			if err := job.Cmd.Process.Signal(syscall.SIGCONT); err != nil {
				return fmt.Errorf("failed to resume job: %v", err)
			}
		}

		job.Status = types.JobStatusRunning
		job.Background = false

		// Wait for the job to complete in foreground
		err := job.Cmd.Wait()
		if err != nil {
			fmt.Printf("Job [%d] exited with error: %v\n", job.ID, err)
		}

		job.Status = types.JobStatusDone
		endTime := time.Now()
		job.EndTime = &endTime
	}

	// Remove from jobs list since it's completed
	delete(jm.jobs, jobID)
	return nil
}

// ResumeInBackground resumes a stopped job in the background
func (jm *JobManager) ResumeInBackground(jobID int) error {
	job, err := jm.GetJob(jobID)
	if err != nil {
		return err
	}

	if job.Status == types.JobStatusDone {
		return fmt.Errorf("job %d has already completed", jobID)
	}

	if job.Status != types.JobStatusStopped {
		return fmt.Errorf("job %d is not stopped", jobID)
	}

	fmt.Printf("Resuming job [%d] in background: %s\n", job.ID, job.Command)

	// Send SIGCONT to resume the process
	if job.Cmd != nil && job.Cmd.Process != nil {
		if err := job.Cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return fmt.Errorf("failed to resume job: %v", err)
		}

		job.Status = types.JobStatusRunning
		job.Background = true
	}

	return nil
}

// KillJob kills a job by sending SIGTERM
func (jm *JobManager) KillJob(jobID int) error {
	job, err := jm.GetJob(jobID)
	if err != nil {
		return err
	}

	if job.Status == types.JobStatusDone {
		return fmt.Errorf("job %d has already completed", jobID)
	}

	fmt.Printf("Terminating job [%d]: %s\n", job.ID, job.Command)

	if job.Cmd != nil && job.Cmd.Process != nil {
		if err := job.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill job: %v", err)
		}

		job.Status = types.JobStatusDone
		endTime := time.Now()
		job.EndTime = &endTime
	}

	return nil
}

// CleanupCompletedJobs removes completed jobs from the manager
func (jm *JobManager) CleanupCompletedJobs() {
	for id, job := range jm.jobs {
		if job.Status == types.JobStatusDone {
			delete(jm.jobs, id)
		}
	}
}
