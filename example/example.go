/*
 * example.go implements an api demo for command flags
 */
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jagipson/commandflags"
)

// A place to put settings
var options = struct {
	verbose   bool
	debug     bool
	cpu       float64
	mem       int
	instances int
}{ // defaults
	verbose:   false,
	debug:     false,
	cpu:       1.0,
	mem:       32,
	instances: 1,
}

// declare global for root commandflags.CommandType object
var cf commandflags.CommandType

func init() {
	// because in this example, all command that have flags have the same flags,
	// we just share the same flagset.

	flags := flag.NewFlagSet("flags", flag.ContinueOnError)
	flags.BoolVar(&options.verbose, "verbose", options.verbose, "Enable verbose output")
	flags.BoolVar(&options.debug, "debug", options.debug, "Enable debug output")
	flags.Float64Var(&options.cpu, "c", options.cpu, "cpu share")
	flags.IntVar(&options.mem, "m", options.mem, "memory share (MB)")
	flags.IntVar(&options.instances, "i", options.instances, "instance count")

	cf = commandflags.NewCommandType("example", flags)
	cf.ShortDesc = `This example demonstrates commandflags.`
	cf.LongDesc = `This example demonstrates commandflags. The commandflags
	library is designed to be a minimal add-on for the built-in flags library to
	add subcommands. Each subcommand may have it's own flag set, or share a
	flagset with some (or all) other subcommands. The heaviest part of the
	commandflags library is really the help system which is improved over the
	default flags help.`
	cf.SubCommands = map[string]commandflags.CommandType{
		"help": commandflags.CommandType{
			Name:      "help",
			ShortDesc: "Show help for a command",
			LongDesc:  "Help usage:  help COMMAND",
			Help:      "Valid commands are deploy, create, update, show, list_artifacts, and deployments",
		},
		"deploy": commandflags.CommandType{
			Name:      "deploy",
			Flags:     flags,
			ShortDesc: "deploy an app completely",
			LongDesc:  "usage: deploy NAME REV",
		},
		"create": commandflags.CommandType{
			Name:      "create",
			Flags:     flags,
			ShortDesc: "initial create/deploy of an app",
			LongDesc:  "usage: create NAME REV",
		},
		"update": commandflags.CommandType{
			Name:      "update",
			Flags:     flags,
			ShortDesc: "update definition of an app, really!",
			LongDesc:  "usage: update NAME REV",
		},
		"show": commandflags.CommandType{
			Name:      "show",
			ShortDesc: "show the description of an app",
		},
		"deployments": commandflags.CommandType{
			Name:      "deployments",
			ShortDesc: "either status of destroy the deployments for an app",
			SubCommands: map[string]commandflags.CommandType{
				"status": commandflags.CommandType{
					Name:      "status",
					ShortDesc: "get the status of deployments for an app",
				},
				"destroy": commandflags.CommandType{
					Name:      "destroy",
					ShortDesc: "destroy hung deployment for an app",
				},
			},
		},
		"list_artifacts": commandflags.CommandType{
			Name:      "list_artifacts",
			ShortDesc: "list_artifacts the definition of an app",
		},
	}
}

func main() {
	var words []string
	var err error

	if words, err = cf.ProcessArgs(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("all the nonflags are: ", words)
}
