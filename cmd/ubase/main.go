package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kernelplex/ubase/internal/commands"
)

func main() {
	// Read dot env
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	commands := commands.GetCommands(os.Args[0])
	err = commands.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
