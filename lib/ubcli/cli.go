package ubcli

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Command struct {
	Name    string
	Help    string
	Run     func(args []string) error
	FlagSet *flag.FlagSet
}

type CommandLine struct {
	commands    []Command
	name        string
	verbose     bool
	globalFlags *flag.FlagSet
}

func NewCommandLine(name string) *CommandLine {
	name = name[strings.LastIndex(name, "/")+1:]
	cl := &CommandLine{
		commands:    make([]Command, 0),
		name:        name,
		globalFlags: flag.NewFlagSet(name, flag.ExitOnError),
	}
	cl.globalFlags.BoolVar(&cl.verbose, "v", false, "Enable verbose output")
	cl.globalFlags.BoolVar(&cl.verbose, "verbose", false, "Enable verbose output")

	cl.Add(Command{
		Name: "help",
		Help: "Show help for a command",
		Run: func(args []string) error {
			cl.Help(args)
			return nil
		},
	})
	return cl
}

func (c *CommandLine) Add(cmd ...Command) {
	for _, command := range cmd {
		c.commands = append(c.commands, command)
	}
}

func (c *CommandLine) Help(args []string) {
	if len(args) == 0 {
		c.Usage(args)
		return
	}

	commandName := args[0]
	for _, c := range c.commands {
		if c.Name == commandName {
			fmt.Printf("Help for %s:\n%s\n", c.Name, c.Help)
			if c.FlagSet != nil {
				fmt.Println("\nFlags:")
				c.FlagSet.PrintDefaults()
			}
			return
		}
	}
}

func (c *CommandLine) Run(args []string) error {
	// Parse global flags first
	err := c.globalFlags.Parse(args[1:])
	if err != nil {
		return err
	}
	args = c.globalFlags.Args() // Get remaining args after global flags

	if len(args) == 0 {
		c.Help(args)
		return nil
	}

	defaultLevel := slog.LevelError
	if c.verbose {
		defaultLevel = slog.LevelDebug
	}

	opts := slog.HandlerOptions{
		Level: defaultLevel,
	}
	handler := slog.NewTextHandler(os.Stderr, &opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	commandName := args[0]

	var invoke *Command = nil
	for _, cmd := range c.commands {
		if cmd.Name == commandName {
			invoke = &cmd
			break
		}
	}

	if invoke == nil {
		fmt.Printf("Unknown command: %s\n", commandName)
		c.Usage(args)
		return nil
	}

	// Prepare command args - skip the command name itself
	cmdArgs := args[1:]

	// If command has its own FlagSet, parse those flags
	if invoke.FlagSet != nil {
		err = invoke.FlagSet.Parse(cmdArgs)
		if err != nil {
			return err
		}
		cmdArgs = invoke.FlagSet.Args() // Get remaining args after command flags
	}

	return invoke.Run(cmdArgs)
}

func (c *CommandLine) Usage(args []string) {
	fmt.Printf("Usage: %s [global-flags] <command> [args]\n\n", c.name)
	fmt.Println("Global flags:")
	c.globalFlags.PrintDefaults()
	fmt.Println("\nAvailable commands:")

	for _, c := range c.commands {
		fmt.Printf("  %-15s %s\n", c.Name, c.Help)
	}

	fmt.Printf("\nGlobal flags:\n")
	flag.PrintDefaults()
}
