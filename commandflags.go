package commandflags

import (
	"bytes"
	"flag"
	"fmt"
)

// CommandType implements a nested Command-flag structure whereby options
// (flags) are processed, and then subcommands are processed. Each subcommand
// is another commandType and the process recurses, each having it's own flag
// set.
type CommandType struct {
	Name        string                 // Name of command
	Flags       *flag.FlagSet          // Flagset for command
	SubCommands map[string]CommandType // map of subcommands
}

// NewCommandType returns an initialized CommandType
func NewCommandType(name string, flags *flag.FlagSet) CommandType {
	return CommandType{Name: name, SubCommands: map[string]CommandType{}, Flags: flags}
}

// The Error interface implements error and adds CommandType() and Args()
// methods that return the CommandType object in which the error occurred and
// the remaining arguments that were being process when the error occurred,
// respectively
type Error interface {
	Error() string             // standard error interface
	CommandType() *CommandType // reference to command type that had error
	Args() []string            // slice of remaining arguments
}

// A UsageError object is defined as the underlying type for the
// MissingCommandError, InvalidCommandError, and FlagError types. It
// implements the Error interface.
type UsageError struct {
	e string       // error message
	c *CommandType // reference to the offended CommandType
	a []string     // the args that offended the CommandType
}

// Error implements the standard error interface in UsageError.
func (e UsageError) Error() string { return e.e }

// CommandType returns the CommandType object that encountered the processing.
// error.
func (e UsageError) CommandType() *CommandType { return e.c }

// Args returns the arguments being process when the error was encountered.
func (e UsageError) Args() []string { return e.a }

// A MissingCommandError is returned when a command expected a sub-command
// (i.e. the CommandType object's SubCommands map was not empty) but there
// were no more arguments remaining to process.
type MissingCommandError struct {
	UsageError
}

// An InvalidCommandError is returned when a command expected a sub-command,
// but the next remaining argument does not match the valid sub-commands in
// the CommandType object's SubCommands map.
type InvalidCommandError struct {
	UsageError
}

// A FlagError is returned when the upstream flag library encounters an error
// while parsing arguments for flags.
type FlagError struct {
	UsageError
}

// ProcessArgs starts the recursive process of setting flags and processing
// sub-commands and returns a slice of strings that correspond to the names of
// the commands/subcommands chosen.
func (c *CommandType) ProcessArgs(args []string) ([]string, Error) {
	// Parse the command line for global opts
	if err := c.Flags.Parse(args); err != nil {
		return []string{c.Name}, FlagError{
			UsageError: UsageError{
				e: err.Error(),
				c: c,
				a: args,
			},
		}
	}
	// remaining arguments after processing flag group
	remaining := c.Flags.Args()

	// If subcommands are defined, then recurse. Otherwise run func()
	if len(c.SubCommands) == 0 {
		return append([]string{c.Name}, remaining...), nil
	}
	if len(remaining) == 0 {
		return []string{c.Name}, MissingCommandError{
			UsageError: UsageError{
				e: fmt.Sprintf("Missing COMMAND:\n%s", c.autoHelp()),
				c: c,
				a: args,
			},
		}
	}
	sc, ok := c.SubCommands[remaining[0]]
	if !ok {
		return []string{c.Name}, InvalidCommandError{
			UsageError: UsageError{
				e: fmt.Sprintf("Invalid COMMAND: %s\n%s", remaining[0], c.autoHelp()),
				c: c,
				a: args,
			},
		}
	}
	cp, err := sc.ProcessArgs(remaining[1:])
	return append([]string{c.Name}, cp...), err
}

// Generates reasonable usage messages for MissingCommandError and
// InvalidCommandError and FlagError objects.
func (c *CommandType) autoHelp() string {
	result := bytes.NewBufferString("")
	result.WriteString(fmt.Sprintf("%s OPTIONS COMMAND\nOPTIONS:\n", c.Name))
	c.Flags.SetOutput(result)
	c.Flags.PrintDefaults()
	result.WriteString(fmt.Sprintf("COMMAND is one of:\n"))
	for k := range c.SubCommands {
		result.WriteString(fmt.Sprintf("%s ", k))
	}
	return result.String()
}
