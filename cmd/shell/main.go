package main

import (
	"flag"
	"fmt"

	"github.com/Su5ubedi/advanced-shell/internal/shell"
)

func main() {
	// Command line flags
	var (
		version = flag.Bool("version", false, "Show version information")
		help    = flag.Bool("help", false, "Show help information")
		debug   = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	if *help {
		showHelp()
		return
	}

	// Create and start the shell
	sh := shell.NewShell()

	if *debug {
		fmt.Println("Debug mode enabled")
		// Additional debug setup could go here
	}

	// Run the shell
	sh.Run()
}

func showVersion() {
	fmt.Println("Advanced Shell Simulation")
	fmt.Println("Version: 1.0.0 (Deliverable 1)")
	fmt.Println("Build: Development")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("- Basic shell functionality")
	fmt.Println("- Built-in commands")
	fmt.Println("- Process management")
	fmt.Println("- Job control")
}

func showHelp() {
	fmt.Println("Advanced Shell Simulation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  shell [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version    Show version information")
	fmt.Println("  -help       Show this help message")
	fmt.Println("  -debug      Enable debug mode")
	fmt.Println()
	fmt.Println("Once started, type 'help' for available shell commands")
}
