package commandflags

import (
	"flag"
	"fmt"

	"github.com/jagipson/refmt"
)

// HelpIndent sets the number of spaces each subcommand's help is indented
var HelpIndent int = 2

// DefaultWidth is the default screen/term width assumed when wrapping text
// for help output.
var DefaultWidth int = 80

// CommandType implements a nested Command-flag structure whereby options
// (flags) are processed, and then subcommands are processed. Each subcommand
// is another commandType and the process recurses, each having it's own flag
// set. The ShortDesc is displayed when the parent command's help lists
// available subcommands. The LongDesc is displayed at the top of the help for
// the command, and the Help is displayed at the bottom. The Help is often
// used to explain the interaction between flags or expand upon flag
// documentation. The flagset will be renamed, to the name of the command, and
// the flag.FlagSet error handling will be reset to flag.ContinueOnError. This
// allows the error handling to be done by commandflags and the downstream
// program.
type CommandType struct {
	Name        string                 // Name of command
	ShortDesc   string                 // Short description of subcommand
	LongDesc    string                 // Detailed description of subcommand
	Help        string                 // Documentation of subcommand
	Flags       *flag.FlagSet          // Flagset for command
	SubCommands map[string]CommandType // map of subcommands
}

// NewCommandType returns an initialized CommandType
func NewCommandType(name string, flags *flag.FlagSet) CommandType {
	c := CommandType{Name: name, SubCommands: map[string]CommandType{}, Flags: flags}
	return c
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
	// reconfigure flags' error handling:
	f := func() {} // noop function
	c.Flags.Init(c.Name, flag.ContinueOnError)
	c.Flags.Usage = f

	// Parse the command line for global opts
	if err := c.Flags.Parse(args); err != nil {
		return []string{c.Name}, FlagError{
			UsageError: UsageError{
				//e: err.Error(),
				e: fmt.Sprintf("%s", c.renderHelp(DefaultWidth)),
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
				e: fmt.Sprintf("Missing COMMAND:\n%s", c.renderHelp(DefaultWidth)),
				c: c,
				a: args,
			},
		}
	}
	sc, ok := c.SubCommands[remaining[0]]
	if !ok {
		return []string{c.Name}, InvalidCommandError{
			UsageError: UsageError{
				e: fmt.Sprintf("Invalid COMMAND: %s\n%s", remaining[0], c.renderHelp(DefaultWidth)),
				c: c,
				a: args,
			},
		}
	}
	cp, err := sc.ProcessArgs(remaining[1:])
	return append([]string{c.Name}, cp...), err
}

func (c CommandType) renderHelp(width int) string {
	style := refmt.NewStyle()
	style.IndentWidth = HelpIndent
	style.MaxWidth = width - HelpIndent
	help := fmt.Sprintf("Command: %s\n", c.Name)

	// Print description, if set -- prefer LongDesc
	switch {
	case len(c.LongDesc) > 0:
		help += fmt.Sprintf("%s\n\n", style.Indent(style.Wrap(c.LongDesc)))
	case len(c.ShortDesc) > 0:
		help += fmt.Sprintf("%s\n\n", style.Indent(style.Wrap(c.ShortDesc)))
	}

	// obtain the flags in the flagset and generate labels
	flags := []*flag.Flag{}
	flagArgs := map[string]string{}
	maxFlagWidth := 0
	appendFlag := func(f *flag.Flag) {
		flags = append(flags, f)
		label := ""
		// Thank frobnitz for figuring this out
		switch f.Value.(flag.Getter).Get().(type) {
		case bool:
			label = ""
		case uint64, uint:
			label = "UINT"
		case int64, int:
			label = "INT"
		case string:
			label = "STRING"
		case float64:
			label = "FLOAT"
		default:
			label = "VALUE"
		}
		flagArgs[f.Name] = label
		if len(f.Name)+len(label) > maxFlagWidth {
			maxFlagWidth = len(f.Name) + len(label)
		}
	}
	c.Flags.VisitAll(appendFlag)

	// set width needed to express flagnames
	flagColWidth := maxFlagWidth + 6 // 4 = 2 for left indent, 1 for dash, 1 for space between name and label, 2 for space at end

	// print help for flags
	flagStyle := refmt.NewStyle()
	flagStyle.MaxWidth = width - flagColWidth
	flagStyle.IndentWidth = flagColWidth
	if len(flags) > 0 {
		help += fmt.Sprintf("%*s%s flags:\n", HelpIndent, "", c.Name)
	}
	for _, f := range flags {
		flag := fmt.Sprintf("%*s-%s %s", HelpIndent, "", f.Name, flagArgs[f.Name])
		help += fmt.Sprintf("%-*s%s\n", flagColWidth, flag, flagStyle.Indent2(flagStyle.Wrap(f.Usage)))
	}

	// exit now if no subcommands
	if len(c.SubCommands) < 1 {
		return help
	}

	help += fmt.Sprintf("\n%*s%s sub-commands:\n", HelpIndent, "", c.Name)
	maxSubcmdWidth := 0
	for _, v := range c.SubCommands {
		if len(v.Name) > maxSubcmdWidth {
			maxSubcmdWidth = len(v.Name)
		}
	}
	cmdStyle := refmt.NewStyle()
	cmdStyle.MaxWidth = width - (HelpIndent + maxSubcmdWidth + 2)
	cmdStyle.IndentWidth = HelpIndent + maxSubcmdWidth + 2
	for _, v := range c.SubCommands {
		help += fmt.Sprintf("%*s%-*s  %s\n", HelpIndent, "", maxSubcmdWidth, v.Name, cmdStyle.Indent2(cmdStyle.Wrap(v.ShortDesc)))
	}
	return help
}
