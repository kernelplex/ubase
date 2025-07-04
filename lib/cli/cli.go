package cli

import (
	"flag"
	"fmt"
	"strings"
)

type Command struct {
	Name    string
	Help    string
	Run     func(args []string) error
	FlagSet *flag.FlagSet
}

type CommandLine struct {
	commands []Command
	name     string
}

func NewCommandLine(name string) *CommandLine {
	name = name[strings.LastIndex(name, "/")+1:]
	cl := &CommandLine{
		commands: make([]Command, 0),
		name:     name,
	}

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
	if len(args) == 1 {
		c.Help(args[1:])
		return nil
	}

	commandName := args[1]

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

	return invoke.Run(args[2:])
}

func (c *CommandLine) Usage(args []string) {

	fmt.Printf("Usage: %s <command> [args]\n\n", c.name)
	fmt.Println("Available commands:")

	for _, c := range c.commands {
		fmt.Printf("  %-15s %s\n", c.Name, c.Help)
	}

	fmt.Printf("\nGlobal flags:\n")
	flag.PrintDefaults()
}
