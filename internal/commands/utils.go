package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	"golang.org/x/term"
)

func GetSystemUsername() string {
	if user, err := user.Current(); err == nil {
		return user.Username
	}
	return "unknown"
}

func GetAgent() string {
	return GetSystemUsername() + "@" + os.Getenv("HOSTNAME")
}

func readLine() (string, error) {
	// Use buffered reader to read a line from standard input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n') // Reads until a newline
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	// Trim the spaces and newline characters
	input = strings.TrimSpace(input)
	return input, nil
}

func maybeReadInput(prompt string, existing string) string {
	if existing != "" {
		return existing
	}

	for true {

		fmt.Print(prompt)
		input, err := readLine()
		if err == nil {
			return input
		}
		fmt.Println("Invalid input: " + err.Error())
	}
	return ""
}

func maybeReadOptionalInput(prompt string, existing string) *string {
	if existing != "" {
		return &existing
	}

	for true {

		fmt.Print(prompt)
		var input string
		input, err := readLine()
		if input == "" && err == nil {
			return nil
		}
		if err == nil {
			return &input
		}
		fmt.Println("Invalid input: " + err.Error())
	}
	return nil
}

func maybeReadPasswordInput(prompt string, value string, required bool) (*string, error) {
	if value != "" {
		return &value, nil
	}
	for {
		// First password entry
		fmt.Print(prompt)
		password, err := readPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}

		// Handle empty password
		if password == "" {
			if !required {
				return nil, nil
			}
			fmt.Println("Password is required")
			continue
		}

		// Password confirmation
		fmt.Print("Confirm password: ")
		confirm, err := readPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if password != confirm {
			fmt.Println("Passwords do not match")
			continue
		}

		return &password, nil
	}
}

func maybeReadInt64Input(prompt string, existing int64) int64 {
	if existing != 0 {
		return existing
	}

	for {
		fmt.Print(prompt)
		var input int64
		_, err := fmt.Scanln(&input)
		if err == nil {
			return input
		}
		fmt.Println("Invalid input: please enter a valid number")
	}
}

func readPassword() (string, error) {
	// For Unix-like systems
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	term := term.NewTerminal(os.Stdin, "")
	return term.ReadPassword("")
}
