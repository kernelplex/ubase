package main

import (
	"log/slog"
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

	opts := slog.HandlerOptions{
		Level: slog.LevelError,
	}
	handler := slog.NewTextHandler(os.Stderr, &opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	commands := commands.GetCommands(os.Args[0])
	commands.Run(os.Args)
}
