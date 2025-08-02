package shell

import (
	"fmt"
	"strings"
)

// CommandParser handles parsing of command line input
type CommandParser struct{}

// NewCommandParser creates a new command parser
func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

// ParsedCommand represents a parsed command
type ParsedCommand struct {
	Command    string
	Args       []string
	Background bool
	Pipes      [][]string // For future pipe implementation
}

// Parse parses a command line input string
func (cp *CommandParser) Parse(input string) *ParsedCommand {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// Check for background execution
	background := false
	if strings.HasSuffix(input, "&") {
		background = true
		input = strings.TrimSpace(strings.TrimSuffix(input, "&"))
	}

	// Simple tokenization (doesn't handle quotes yet)
	args := cp.tokenize(input)
	if len(args) == 0 {
		return nil
	}

	return &ParsedCommand{
		Command:    args[0],
		Args:       args,
		Background: background,
		Pipes:      [][]string{args}, // Single command for now
	}
}

// tokenize splits input into tokens
func (cp *CommandParser) tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	for _, char := range input {
		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(char)
			}
		case char == ' ' || char == '\t':
			if inQuotes {
				current.WriteRune(char)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last token if there is one
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// IsBuiltinCommand checks if a command is a built-in command
func (cp *CommandParser) IsBuiltinCommand(command string) bool {
	builtins := map[string]bool{
		"cd":    true,
		"pwd":   true,
		"exit":  true,
		"echo":  true,
		"clear": true,
		"ls":    true,
		"cat":   true,
		"mkdir": true,
		"rmdir": true,
		"rm":    true,
		"touch": true,
		"kill":  true,
		"jobs":  true,
		"fg":    true,
		"bg":    true,
		"help":  true,
	}

	return builtins[command]
}

// ValidateCommand performs comprehensive validation on parsed commands
func (cp *CommandParser) ValidateCommand(parsed *ParsedCommand) error {
	if parsed == nil || parsed.Command == "" {
		return nil // Empty command is valid (just ignored)
	}

	// Check for dangerous command patterns
	if strings.Contains(parsed.Command, "..") {
		return fmt.Errorf("potentially dangerous path detected: %s", parsed.Command)
	}

	// Validate command name (no special characters except allowed ones)
	if strings.ContainsAny(parsed.Command, "|;&<>(){}[]") {
		return fmt.Errorf("invalid characters in command name: %s", parsed.Command)
	}

	// Check for excessively long commands
	if len(parsed.Command) > 256 {
		return fmt.Errorf("command name too long (max 256 characters)")
	}

	// Validate arguments
	for i, arg := range parsed.Args {
		if len(arg) > 1024 {
			return fmt.Errorf("argument %d too long (max 1024 characters)", i)
		}
	}

	// Check total argument count
	if len(parsed.Args) > 100 {
		return fmt.Errorf("too many arguments (max 100)")
	}

	return nil
}
