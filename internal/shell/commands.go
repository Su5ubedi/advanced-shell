package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CommandHandler handles built-in shell commands
type CommandHandler struct {
	jobManager *JobManager
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(jobManager *JobManager) *CommandHandler {
	return &CommandHandler{
		jobManager: jobManager,
	}
}

// HandleCommand executes a built-in command
func (ch *CommandHandler) HandleCommand(parsed *ParsedCommand) error {
	if parsed == nil || parsed.Command == "" {
		return nil
	}

	switch parsed.Command {
	case "cd":
		return ch.handleCD(parsed.Args)
	case "pwd":
		return ch.handlePWD(parsed.Args)
	case "exit":
		return ch.handleExit(parsed.Args)
	case "echo":
		return ch.handleEcho(parsed.Args)
	case "clear":
		return ch.handleClear(parsed.Args)
	case "ls":
		return ch.handleLS(parsed.Args)
	case "cat":
		return ch.handleCat(parsed.Args)
	case "mkdir":
		return ch.handleMkdir(parsed.Args)
	case "rmdir":
		return ch.handleRmdir(parsed.Args)
	case "rm":
		return ch.handleRm(parsed.Args)
	case "touch":
		return ch.handleTouch(parsed.Args)
	case "kill":
		return ch.handleKill(parsed.Args)
	case "jobs":
		return ch.handleJobs(parsed.Args)
	case "fg":
		return ch.handleFG(parsed.Args)
	case "bg":
		return ch.handleBG(parsed.Args)
	case "help":
		return ch.handleHelp(parsed.Args)
	default:
		return fmt.Errorf("unknown built-in command: %s", parsed.Command)
	}
}

func (ch *CommandHandler) handleCD(args []string) error {
	var dir string
	if len(args) < 2 {
		// Change to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: cannot determine home directory: %v", err)
		}
		dir = homeDir
	} else if len(args) > 2 {
		return fmt.Errorf("cd: too many arguments")
	} else {
		dir = args[1]

		// Validate directory argument
		if dir == "" {
			return fmt.Errorf("cd: empty directory name")
		}

		// Handle special cases
		if dir == "-" {
			// TODO: Implement previous directory functionality
			return fmt.Errorf("cd: previous directory functionality not implemented yet")
		}
		if dir == "~" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cd: cannot determine home directory: %v", err)
			}
			dir = homeDir
		}
		if strings.HasPrefix(dir, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cd: cannot determine home directory: %v", err)
			}
			dir = filepath.Join(homeDir, dir[2:])
		}
	}

	// Check if directory exists before trying to change
	if stat, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cd: %s: no such file or directory", dir)
		} else if os.IsPermission(err) {
			return fmt.Errorf("cd: %s: permission denied", dir)
		}
		return fmt.Errorf("cd: %s: %v", dir, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("cd: %s: not a directory", dir)
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %s: %v", dir, err)
	}
	return nil
}

func (ch *CommandHandler) handlePWD(args []string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %v", err)
	}
	fmt.Println(pwd)
	return nil
}

func (ch *CommandHandler) handleExit(args []string) error {
	fmt.Println("Goodbye!")

	// Clean shutdown - kill any remaining jobs
	jobs := ch.jobManager.GetAllJobs()
	for _, job := range jobs {
		if job.Status != "Done" {
			fmt.Printf("Terminating job [%d]: %s\n", job.ID, job.Command)
			ch.jobManager.KillJob(job.ID)
		}
	}

	os.Exit(0)
	return nil
}

func (ch *CommandHandler) handleEcho(args []string) error {
	if len(args) > 1 {
		// Join all arguments except the command itself
		output := strings.Join(args[1:], " ")

		// Handle basic escape sequences
		output = strings.ReplaceAll(output, "\\n", "\n")
		output = strings.ReplaceAll(output, "\\t", "\t")

		fmt.Println(output)
	}
	return nil
}

func (ch *CommandHandler) handleClear(args []string) error {
	// Try different clear commands based on OS
	var cmd *exec.Cmd

	// Check if we're on Windows
	if _, err := exec.LookPath("cls"); err == nil {
		cmd = exec.Command("cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (ch *CommandHandler) handleLS(args []string) error {
	dir := "."
	showHidden := false
	longFormat := false
	var invalidFlags []string

	// Parse flags and directory
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// Handle flags
			for _, flag := range arg[1:] {
				switch flag {
				case 'a':
					showHidden = true
				case 'l':
					longFormat = true
				default:
					invalidFlags = append(invalidFlags, string(flag))
				}
			}
		} else {
			if dir != "." {
				return fmt.Errorf("ls: too many directory arguments")
			}
			dir = arg
		}
	}

	// Report invalid flags
	if len(invalidFlags) > 0 {
		return fmt.Errorf("ls: invalid option(s): %s", strings.Join(invalidFlags, ", "))
	}

	// Check if directory exists and is accessible
	if stat, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ls: %s: no such file or directory", dir)
		} else if os.IsPermission(err) {
			return fmt.Errorf("ls: %s: permission denied", dir)
		}
		return fmt.Errorf("ls: %s: %v", dir, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("ls: %s: not a directory", dir)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ls: %s: %v", dir, err)
	}

	for _, file := range files {
		// Skip hidden files unless -a flag is used
		if !showHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if longFormat {
			info, err := file.Info()
			if err != nil {
				fmt.Printf("? %s\n", file.Name())
				continue
			}

			mode := info.Mode()
			modTime := info.ModTime().Format("Jan 02 15:04")
			size := info.Size()

			fmt.Printf("%s %8d %s %s\n", mode.String(), size, modTime, file.Name())
		} else {
			if file.IsDir() {
				fmt.Printf("%s/\n", file.Name())
			} else {
				fmt.Println(file.Name())
			}
		}
	}
	return nil
}

func (ch *CommandHandler) handleCat(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("cat: missing filename\nUsage: cat [file1] [file2] ...")
	}

	for i := 1; i < len(args); i++ {
		filename := args[i]

		// Validate filename
		if filename == "" {
			return fmt.Errorf("cat: empty filename")
		}

		// Check if file exists and is readable
		if stat, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("cat: %s: no such file or directory", filename)
			} else if os.IsPermission(err) {
				return fmt.Errorf("cat: %s: permission denied", filename)
			}
			return fmt.Errorf("cat: %s: %v", filename, err)
		} else if stat.IsDir() {
			return fmt.Errorf("cat: %s: is a directory", filename)
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("cat: %s: %v", filename, err)
		}
		fmt.Print(string(content))
	}
	return nil
}

func (ch *CommandHandler) handleMkdir(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("mkdir: missing directory name")
	}

	createParents := false
	var dirs []string

	// Parse flags and directories
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "p") {
				createParents = true
			}
		} else {
			dirs = append(dirs, arg)
		}
	}

	if len(dirs) == 0 {
		return fmt.Errorf("mkdir: missing directory name")
	}

	for _, dirname := range dirs {
		var err error
		if createParents {
			err = os.MkdirAll(dirname, 0755)
		} else {
			err = os.Mkdir(dirname, 0755)
		}

		if err != nil {
			return fmt.Errorf("mkdir: %s: %v", dirname, err)
		}
	}
	return nil
}

func (ch *CommandHandler) handleRmdir(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("rmdir: missing directory name")
	}

	for i := 1; i < len(args); i++ {
		dirname := args[i]
		if err := os.Remove(dirname); err != nil {
			return fmt.Errorf("rmdir: %s: %v", dirname, err)
		}
	}
	return nil
}

func (ch *CommandHandler) handleRm(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("rm: missing filename")
	}

	recursive := false
	force := false
	var files []string

	// Parse flags and files
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "r") || strings.Contains(arg, "R") {
				recursive = true
			}
			if strings.Contains(arg, "f") {
				force = true
			}
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("rm: missing filename")
	}

	for _, filename := range files {
		var err error
		if recursive {
			err = os.RemoveAll(filename)
		} else {
			err = os.Remove(filename)
		}

		if err != nil && !force {
			return fmt.Errorf("rm: %s: %v", filename, err)
		}
	}
	return nil
}

func (ch *CommandHandler) handleTouch(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("touch: missing filename")
	}

	for i := 1; i < len(args); i++ {
		filename := args[i]

		// Check if file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// Create the file
			file, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("touch: %s: %v", filename, err)
			}
			file.Close()
		} else {
			// Update timestamp
			now := time.Now()
			if err := os.Chtimes(filename, now, now); err != nil {
				return fmt.Errorf("touch: %s: %v", filename, err)
			}
		}
	}
	return nil
}

func (ch *CommandHandler) handleKill(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("kill: missing PID\nUsage: kill [pid1] [pid2] ...")
	}

	var errors []string
	killed := 0

	for i := 1; i < len(args); i++ {
		pidStr := args[i]

		// Validate PID format
		if pidStr == "" {
			errors = append(errors, "empty PID")
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid PID '%s': not a number", pidStr))
			continue
		}

		// Validate PID range
		if pid <= 0 {
			errors = append(errors, fmt.Sprintf("invalid PID %d: must be positive", pid))
			continue
		}

		// Don't allow killing init process or shell itself
		if pid == 1 {
			errors = append(errors, "cannot kill init process (PID 1)")
			continue
		}

		if pid == os.Getpid() {
			errors = append(errors, "cannot kill shell process itself")
			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			errors = append(errors, fmt.Sprintf("process %d not found: %v", pid, err))
			continue
		}

		if err := process.Kill(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to kill process %d: %v", pid, err))
			continue
		}

		fmt.Printf("Process %d killed\n", pid)
		killed++
	}

	// Report any errors
	if len(errors) > 0 {
		if killed == 0 {
			return fmt.Errorf("kill: %s", strings.Join(errors, "; "))
		} else {
			fmt.Printf("kill: warnings: %s\n", strings.Join(errors, "; "))
		}
	}

	return nil
}

func (ch *CommandHandler) handleJobs(args []string) error {
	ch.jobManager.CleanupCompletedJobs()
	ch.jobManager.ListJobs()
	return nil
}

func (ch *CommandHandler) handleFG(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("fg: missing job ID\nUsage: fg [job_id]\nUse 'jobs' to see available jobs")
	}

	if len(args) > 2 {
		return fmt.Errorf("fg: too many arguments")
	}

	jobIDStr := args[1]
	if jobIDStr == "" {
		return fmt.Errorf("fg: empty job ID")
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		return fmt.Errorf("fg: invalid job ID '%s': not a number", jobIDStr)
	}

	if jobID <= 0 {
		return fmt.Errorf("fg: invalid job ID %d: must be positive", jobID)
	}

	// Check if any jobs exist
	allJobs := ch.jobManager.GetAllJobs()
	if len(allJobs) == 0 {
		return fmt.Errorf("fg: no jobs to bring to foreground")
	}

	return ch.jobManager.BringToForeground(jobID)
}

func (ch *CommandHandler) handleBG(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("bg: missing job ID\nUsage: bg [job_id]\nUse 'jobs' to see available jobs")
	}

	if len(args) > 2 {
		return fmt.Errorf("bg: too many arguments")
	}

	jobIDStr := args[1]
	if jobIDStr == "" {
		return fmt.Errorf("bg: empty job ID")
	}

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		return fmt.Errorf("bg: invalid job ID '%s': not a number", jobIDStr)
	}

	if jobID <= 0 {
		return fmt.Errorf("bg: invalid job ID %d: must be positive", jobID)
	}

	// Check if any jobs exist
	allJobs := ch.jobManager.GetAllJobs()
	if len(allJobs) == 0 {
		return fmt.Errorf("bg: no jobs to resume in background")
	}

	return ch.jobManager.ResumeInBackground(jobID)
}

func (ch *CommandHandler) handleHelp(args []string) error {
	fmt.Println("Advanced Shell - Available Commands:")
	fmt.Println()
	fmt.Println("Built-in Commands:")
	fmt.Println("  cd [directory]     - Change directory (supports ~, -, and relative paths)")
	fmt.Println("  pwd               - Print working directory")
	fmt.Println("  echo [text]       - Print text (supports \\n, \\t escape sequences)")
	fmt.Println("  clear             - Clear screen")
	fmt.Println("  ls [options] [dir] - List files (-a for hidden, -l for long format)")
	fmt.Println("  cat [files...]    - Display file contents")
	fmt.Println("  mkdir [options] [dirs...] - Create directories (-p for parents)")
	fmt.Println("  rmdir [dirs...]   - Remove empty directories")
	fmt.Println("  rm [options] [files...] - Remove files (-r recursive, -f force)")
	fmt.Println("  touch [files...]  - Create empty files or update timestamps")
	fmt.Println("  kill [pids...]    - Kill processes by PID")
	fmt.Println("  exit              - Exit shell")
	fmt.Println("  help              - Show this help")
	fmt.Println()
	fmt.Println("Job Control:")
	fmt.Println("  jobs              - List background jobs")
	fmt.Println("  fg [job_id]       - Bring job to foreground")
	fmt.Println("  bg [job_id]       - Resume job in background")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  command &         - Run command in background")
	fmt.Println("  Ctrl+C            - Interrupt current foreground process")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ls -la")
	fmt.Println("  mkdir -p path/to/dir")
	fmt.Println("  rm -rf unwanted_dir")
	fmt.Println("  sleep 10 &")
	fmt.Println("  jobs")
	fmt.Println("  fg 1")
	fmt.Println("  cat file1.txt file2.txt")
	fmt.Println("  echo \"Hello\\nWorld\"")
	fmt.Println()
	fmt.Println("Advanced Features (Future Deliverables):")
	fmt.Println("  - Process scheduling algorithms")
	fmt.Println("  - Memory management simulation")
	fmt.Println("  - Process synchronization")
	fmt.Println("  - Command piping")
	fmt.Println("  - User authentication and file permissions")
	fmt.Println()
	return nil
}
