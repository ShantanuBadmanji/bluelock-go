package jobscheduler

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bluelock-go/config"
	"github.com/bluelock-go/shared"
	"github.com/bluelock-go/shared/storage/state/statemanager"
	"github.com/robfig/cron/v3"
)

type JobScheduler struct {
	logger       *shared.CustomLogger
	stateManager *statemanager.StateManager
	JobName      string
	job          func() error
	config       *config.Config
}

func NewJobScheduler(customLogger *shared.CustomLogger, stateManager *statemanager.StateManager, jobName string, job func() error,
	config *config.Config) (*JobScheduler, error) {
	if customLogger == nil {
		return nil, fmt.Errorf("custom logger is nil")
	}

	return &JobScheduler{
		logger:       customLogger,
		stateManager: stateManager,
		JobName:      jobName,
		job:          job,
		config:       config,
	}, nil
}

func (js *JobScheduler) Run() {
	js.logger.Info(fmt.Sprintf("Running the job: %s", js.JobName))

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle shutdown in a separate goroutine
	go func() {
		<-sigChan

		js.logger.Info("Received shutdown signal. Saving state...")
		if err := js.stateManager.SaveStateWithMutex(); err != nil {
			js.logger.Error("Failed to save state before shutdown", "error", err)
			js.logger.Error("Exiting without saving state")
			os.Exit(1)
		} else {
			js.logger.Info("State saved successfully before shutdown")
			js.logger.Info("Exiting gracefully")
			os.Exit(0)
		}
	}()

	for {
		// Parse the cron expression
		schedule, err := cron.ParseStandard(js.config.Common.CronExpression)
		if err != nil {
			js.logger.Error("Invalid cron expression", "error", err)
			return
		}

		// Calculate the next run time based on the cron expression if not the first run
		if !js.stateManager.State.LastJobExecutionEndTime.IsZero() {
			now := time.Now()
			nextRun := schedule.Next(now)
			js.logger.Info("Next job scheduled", "time", nextRun.Format(time.RFC3339))

			// Sleep until the next scheduled time
			sleepDuration := time.Until(nextRun)
			time.Sleep(sleepDuration)
		}

		// Start the job
		js.stateManager.UpdateOngoingJobStartTime(time.Now())
		js.logger.Info("Job started", "time", time.Now().Format(time.RFC3339))

		err = js.job()

		js.stateManager.UpdateLastJobExecutionTime(time.Now())
		js.logger.Info("Job completed", "time", time.Now().Format(time.RFC3339))

		if err != nil {
			js.logger.Error("Job execution failed: job", "jobName", js.JobName, "error", err)
			os.Exit(1)
		} else {
			js.logger.Info("Job execution completed successfully", "jobName", js.JobName)
		}
	}
}
