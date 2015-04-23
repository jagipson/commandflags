# commandflags
A bolt-on extension to the Go flag library that implements a tree of  COMMAND + FLAGS + SUBCOMMAND + FLAGS ......

Each command may have its own flag.Flagset, and both short and long descriptions that are used when writing help. If a command has sub commands, then the subcommand must also be specified. The flags for a command immediately follow the command.
