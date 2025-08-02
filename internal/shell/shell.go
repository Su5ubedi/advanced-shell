package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Shell represents the main shell instance
type Shell struct {
	jobManager     *JobManager
	commandHandler *CommandHandler
	parser         *CommandParser
	running        bool
	prompt         string
}

// NewShell creates a new shell instance
func NewShell() *Shell {
	jobManager := NewJobManager()
	commandHandler := NewCommandHandler(jobManager)
	parser := NewCommandParser()

	return &Shell{
		jobManager:     jobManager,
		commandHandler: commandHandler,
		parser:         parser,
		running:        true,
		prompt:         "[shell]$ ",
	}
}

// Run starts the main shell loop
func (s *Shell) Run() {
	s.setupSignalHandlers()
	s.printWelcome()

	scanner := bufio.NewScanner(os.Stdin)

	for s.running {
		s.displayPrompt()

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		// Handle empty input
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Process the input and handle errors gracefully
		if err := s.processInput(input); err != nil {
			// Color-coded error messages
			fmt.Printf("\033[31mError:\033[0m %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Printf("\033[31mError reading input:\033[0m %v\n", err)
	}

	s.shutdown()
}

// processInput processes a single line of input
func (s *Shell) processInput(input string) error {
	// Parse the command
	parsed := s.parser.Parse(input)
	if parsed == nil {
		return nil // Empty command
	}

	// Validate the command
	if err := s.parser.ValidateCommand(parsed); err != nil {
		return err
	}

	// Check if it's a built-in command
	if s.parser.IsBuiltinCommand(parsed.Command) {
		return s.commandHandler.HandleCommand(parsed)
	}

	// Check if external command exists before trying to execute
	if _, err := exec.LookPath(parsed.Command); err != nil {
		return fmt.Errorf("%s: command not found", parsed.Command)
	}

	return fmt.Errorf("%s: command not found (only built-in commands are supported)", parsed.Command)
}

// setupSignalHandlers sets up signal handlers for graceful shutdown
func (s *Shell) setupSignalHandlers() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nReceived interrupt signal. Use 'exit' to quit the shell.")
		// Don't exit immediately, let user decide
	}()
}

// displayPrompt shows the shell prompt
func (s *Shell) displayPrompt() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Print(s.prompt)
		return
	}

	// Show current directory in prompt
	dir := filepath.Base(pwd)
	if dir == "." {
		dir = pwd
	}

	// Get current time for enhanced prompt
	now := time.Now()
	timeStr := now.Format("15:04:05")

	fmt.Printf("[shell:%s %s]$ ", dir, timeStr)
}

// printWelcome prints the welcome message
func (s *Shell) printWelcome() {
	fmt.Println("==========================================")
	fmt.Println("  Advanced Shell Simulation - Deliverable 1")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Features implemented:")
	fmt.Println("✓ Built-in commands (cd, pwd, ls, cat, etc.)")
	fmt.Println("✓ Process management (foreground/background)")
	fmt.Println("✓ Job control (jobs, fg, bg)")
	fmt.Println("✓ Signal handling")
	fmt.Println("✓ Error handling")
	fmt.Println()
	fmt.Println("Type 'help' for available commands")
	fmt.Println("Type 'exit' to quit")
	fmt.Println()
}

// shutdown performs cleanup before exiting
func (s *Shell) shutdown() {
	fmt.Println("\nShutting down shell...")

	// Get all active jobs
	jobs := s.jobManager.GetAllJobs()
	if len(jobs) > 0 {
		fmt.Printf("Terminating %d active job(s)...\n", len(jobs))

		for _, job := range jobs {
			if job.Status != "Done" {
				fmt.Printf("Killing job [%d]: %s\n", job.ID, job.Command)
				s.jobManager.KillJob(job.ID)
			}
		}

		// Give processes time to terminate
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Goodbye!")
}
